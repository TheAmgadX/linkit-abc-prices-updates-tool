package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/internal"
	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/internal/excel"
)

const (
	reset = "\033[0m"
	bold  = "\033[1m"
	green = "\033[32m"
	red   = "\033[31m"
	cyan  = "\033[36m"
	gray  = "\033[90m"
)

func main() {
	stockFile := flag.String(
		"stock",
		"",
		"Path to the Stock Excel file",
	)

	imfFile := flag.String(
		"imf",
		"",
		"Path to the IMF Excel file",
	)

	outputFile := flag.String(
		"output",
		"",
		"Path to the output Excel file",
	)

	flag.Parse()

	printBanner()

	reader := bufio.NewReader(os.Stdin)

	// If flags weren't provided, ask interactively.
	if *stockFile == "" {
		*stockFile = prompt(
			reader,
			"Path for Stock Excel file",
		)
	}

	if *imfFile == "" {
		*imfFile = prompt(
			reader,
			"Path for IMF Excel file",
		)
	}

	if *outputFile == "" {
		*outputFile = prompt(
			reader,
			"Path for output Excel file with file name with extension",
		)
	}

	fmt.Println()

	fmt.Printf("%sStock:%s  %s\n", cyan, reset, *stockFile)
	fmt.Printf("%sIMF:%s    %s\n", cyan, reset, *imfFile)
	fmt.Printf("%sOutput:%s  %s\n\n", cyan, reset, *outputFile)

	// ---------------------------------------------------------
	// Create file managers
	// ---------------------------------------------------------

	stockManager := excel.NewExcelFileManager(
		*stockFile,
		"",
		[]string{"ProductId", "SalePrice"},
	)

	imfManager := excel.NewExcelFileManager(
		*imfFile,
		"",
		[]string{"ProductId", "VAT"},
	)

	// ---------------------------------------------------------
	// Read both files in parallel
	// ---------------------------------------------------------

	fmt.Printf("%s[%s]%s Reading input files...\n", cyan, "*", reset)

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	var stockErr error
	var imfErr error

	go func() {
		defer wg.Done()
		stockErr = stockManager.Read()
	}()

	go func() {
		defer wg.Done()
		imfErr = imfManager.Read()
	}()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

loop:
	for {
		select {
		case <-done:
			break loop

		default:
			fmt.Printf(
				"\r%s%s%s Reading... ",
				cyan,
				spinner[i%len(spinner)],
				reset,
			)

			i++
			time.Sleep(80 * time.Millisecond)
		}
	}

	fmt.Print("\r\033[K")

	if stockErr != nil {
		fail("Failed to read Stock file", stockErr)
	}

	if imfErr != nil {
		fail("Failed to read IMF file", imfErr)
	}

	fmt.Printf(
		"%s✓%s Stock: %s%d%s products\n",
		green,
		reset,
		bold,
		len(stockManager.Data),
		reset,
	)

	fmt.Printf(
		"%s✓%s IMF:   %s%d%s products\n",
		green,
		reset,
		bold,
		len(imfManager.Data),
		reset,
	)

	fmt.Printf(
		"%s✓%s Read completed in %s%.2fs%s\n\n",
		green,
		reset,
		bold,
		time.Since(start).Seconds(),
		reset,
	)

	// ---------------------------------------------------------
	// Processing
	// ---------------------------------------------------------

	fmt.Printf("%s[%s]%s Processing...\n", cyan, "*", reset)

	processingStart := time.Now()

	processor := internal.NewProcessor(
		&stockManager.Data,
		&imfManager.Data,
	)

	result := processor.Process()

	fmt.Printf(
		"%s✓%s Processed: %s%d%s products\n",
		green,
		reset,
		bold,
		len(result),
		reset,
	)

	fmt.Printf(
		"%s✓%s Processing completed in %s%.2fs%s\n\n",
		green,
		reset,
		bold,
		time.Since(processingStart).Seconds(),
		reset,
	)

	// ---------------------------------------------------------
	// Writing
	// ---------------------------------------------------------

	fmt.Printf("%s[%s]%s Writing output...\n", cyan, "*", reset)

	writer := excel.NewExcelFileManager(
		"",
		*outputFile,
		nil,
	)

	writer.Data = result

	writeStart := time.Now()

	if err := writer.WriteStream(); err != nil {
		fail("Failed to write output file", err)
	}

	fmt.Printf(
		"%s✓%s Output written successfully\n",
		green,
		reset,
	)

	fmt.Printf(
		"%s✓%s Writing completed in %s%.2fs%s\n\n",
		green,
		reset,
		bold,
		time.Since(writeStart).Seconds(),
		reset,
	)

	// ---------------------------------------------------------
	// Summary
	// ---------------------------------------------------------

	fmt.Println(
		gray + "────────────────────────────────────────────" + reset,
	)

	fmt.Printf(
		"%s%sDone!%s\n",
		bold,
		green,
		reset,
	)

	fmt.Printf(
		"%sOutput:%s %s\n",
		cyan,
		reset,
		*outputFile,
	)

	fmt.Printf(
		"%sTotal:%s %.2fs\n",
		cyan,
		reset,
		time.Since(start).Seconds(),
	)
}

func prompt(reader *bufio.Reader, name string) string {
	for {
		fmt.Printf(
			"%s?%s %s%s%s: ",
			cyan,
			reset,
			bold,
			name,
			reset,
		)

		value, err := reader.ReadString('\n')
		if err != nil {
			fail("Failed to read input", err)
		}

		value = strings.TrimSpace(value)

		if value != "" {
			return value
		}

		fmt.Printf(
			"%s✗%s Value cannot be empty.\n",
			red,
			reset,
		)
	}
}

func printBanner() {
	fmt.Printf(`
%s%s
╔══════════════════════════════════════════════╗
║                                              ║
║        LINKIT • ABC PRICE UPDATER            ║
║                                              ║
║        Stock × IMF → Price Updates           ║
║                                              ║
╚══════════════════════════════════════════════╝
%s
`, bold, cyan, reset)
}

func fail(message string, err error) {
	fmt.Printf(
		"%s✗%s %s: %v\n",
		red,
		reset,
		message,
		err,
	)

	os.Exit(1)
}
