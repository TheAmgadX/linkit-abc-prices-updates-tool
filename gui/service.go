package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/internal"
	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/internal/excel"
)

// service is the application service that orchestrates the full processing pipeline.
// It coordinates concurrent file reading, delegates all business logic to the
// existing internal packages, and writes output reports.
type service struct{}

var appService = &service{}

// sanitizeOrgName makes the organisation name safe to use inside a filename.
// Runs of non-alphanumeric characters are collapsed into a single hyphen.
func sanitizeOrgName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := re.ReplaceAllString(strings.TrimSpace(name), "-")
	sanitized = strings.Trim(sanitized, "-")
	return sanitized
}

// uniquePath returns a path that does not yet exist on disk.
// If the given path is free it is returned unchanged. Otherwise a numeric
// suffix is appended before the extension until a free slot is found.
func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; i <= 999; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}

	// Extremely unlikely — fall back to nanosecond timestamp suffix.
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

// Process runs the complete processing pipeline and returns a summary DTO.
// All actual business logic is performed by the existing internal packages.
func (s *service) Process(input ProcessInput) ProcessingResult {
	start := time.Now()

	// ---- Validate inputs -------------------------------------------------
	if err := validateInput(input); err != nil {
		return ProcessingResult{Error: err.Error()}
	}

	// ---- Build output file paths -----------------------------------------
	orgSlug := sanitizeOrgName(input.OrgName)
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	mainFile := uniquePath(filepath.Join(
		input.OutputDir,
		fmt.Sprintf("%s-price-updates-%s.xlsx", orgSlug, timestamp),
	))
	missingInIMFFile := filepath.Join(
		input.OutputDir,
		fmt.Sprintf("%s-iv_ids-in-stock-missing-in-imf-%s.xlsx", orgSlug, timestamp),
	)
	missingInStockFile := filepath.Join(
		input.OutputDir,
		fmt.Sprintf("%s-iv_ids-in-imf-missing-in-stock-%s.xlsx", orgSlug, timestamp),
	)

	// ---- Read Stock and IMF concurrently ---------------------------------
	stockManager := excel.NewExcelFileManager(
		input.StockFilePath,
		"",
		[]string{input.StockIVIDColumn, input.StockPriceColumn},
		input.StockPriceColumn,
		input.StockIVIDColumn,
		"",
	)
	imfManager := excel.NewExcelFileManager(
		input.IMFFilePath,
		"",
		[]string{input.IMFIVIDColumn, input.IMFVATColumn},
		"",
		input.IMFIVIDColumn,
		input.IMFVATColumn,
	)

	var wg sync.WaitGroup
	wg.Add(2)

	var stockErr, imfErr error

	go func() {
		defer wg.Done()
		stockErr = stockManager.Read()
	}()

	go func() {
		defer wg.Done()
		imfErr = imfManager.Read()
	}()

	wg.Wait()

	if stockErr != nil {
		return ProcessingResult{Error: fmt.Sprintf("Failed to read Stock file: %v", stockErr)}
	}
	if imfErr != nil {
		return ProcessingResult{Error: fmt.Sprintf("Failed to read IMF file: %v", imfErr)}
	}

	stockCount := len(stockManager.Data)
	imfCount := len(imfManager.Data)

	// ---- Process ---------------------------------------------------------
	processor := internal.NewProcessor(&stockManager.Data, &imfManager.Data)
	result := processor.Process()

	processedCount := len(result)
	missingInIMFCount := len(processor.NotFoundIVID)
	missingInStockCount := len(processor.HaveNoMatchingIVID)

	// ---- Write main report -----------------------------------------------
	writer := excel.NewExcelFileManager("", mainFile, nil, "", "", "")
	writer.Data = result

	if err := writer.WriteStream(); err != nil {
		return ProcessingResult{Error: fmt.Sprintf("Failed to write main report: %v", err)}
	}

	outputFiles := []string{filepath.Base(mainFile)}

	// ---- Write mismatch reports conditionally ----------------------------
	if missingInIMFCount > 0 {
		missingInIMFFile = uniquePath(missingInIMFFile)
		reportWriter := excel.NewExcelFileManager("", missingInIMFFile, nil, "", "", "")
		if err := reportWriter.WriteStreamReport(processor.NotFoundIVID); err != nil {
			return ProcessingResult{Error: fmt.Sprintf("Failed to write missing-in-IMF report: %v", err)}
		}
		outputFiles = append(outputFiles, filepath.Base(missingInIMFFile))
	}

	if missingInStockCount > 0 {
		missingInStockFile = uniquePath(missingInStockFile)
		reportWriter := excel.NewExcelFileManager("", missingInStockFile, nil, "", "", "")
		if err := reportWriter.WriteStreamReport(processor.HaveNoMatchingIVID); err != nil {
			return ProcessingResult{Error: fmt.Sprintf("Failed to write missing-in-Stock report: %v", err)}
		}
		outputFiles = append(outputFiles, filepath.Base(missingInStockFile))
	}

	durationMs := time.Since(start).Milliseconds()

	return ProcessingResult{
		ProductsProcessed: processedCount,
		StockProducts:     stockCount,
		IMFProducts:       imfCount,
		MissingInIMF:      missingInIMFCount,
		MissingInStock:    missingInStockCount,
		OutputFiles:       outputFiles,
		DurationMs:        durationMs,
		OutputDir:         input.OutputDir,
	}
}

// validateInput checks that all required fields are present and the paths exist.
func validateInput(input ProcessInput) error {
	if strings.TrimSpace(input.OrgName) == "" {
		return fmt.Errorf("organisation name is required")
	}
	if strings.TrimSpace(input.OutputDir) == "" {
		return fmt.Errorf("output directory is required")
	}
	if strings.TrimSpace(input.IMFFilePath) == "" {
		return fmt.Errorf("IMF file is required")
	}
	if strings.TrimSpace(input.StockFilePath) == "" {
		return fmt.Errorf("stock file is required")
	}
	if strings.TrimSpace(input.StockIVIDColumn) == "" {
		return fmt.Errorf("stock Price ID column name is required")
	}
	if strings.TrimSpace(input.StockPriceColumn) == "" {
		return fmt.Errorf("stock Price column name is required")
	}
	if strings.TrimSpace(input.IMFIVIDColumn) == "" {
		return fmt.Errorf("IMF Price ID column name is required")
	}
	if strings.TrimSpace(input.IMFVATColumn) == "" {
		return fmt.Errorf("IMF VAT column name is required")
	}

	if info, err := os.Stat(input.OutputDir); err != nil || !info.IsDir() {
		return fmt.Errorf("output directory does not exist or is not a directory: %s", input.OutputDir)
	}
	if _, err := os.Stat(input.IMFFilePath); err != nil {
		return fmt.Errorf("IMF file not found: %s", input.IMFFilePath)
	}
	if _, err := os.Stat(input.StockFilePath); err != nil {
		return fmt.Errorf("Stock file not found: %s", input.StockFilePath)
	}

	return nil
}
