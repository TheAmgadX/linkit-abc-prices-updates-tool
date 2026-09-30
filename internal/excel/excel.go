package excel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/internal"
	excel "github.com/xuri/excelize/v2"
)

type ExcelFileManager struct {
	Data       map[internal.IVID]internal.Product
	InputFile  string
	OutputFile string
	Columns    []string
}

func NewExcelFileManager(inputFile, outputFile string, columns []string) *ExcelFileManager {
	return &ExcelFileManager{
		InputFile:  inputFile,
		OutputFile: outputFile,
		Columns:    columns,
	}
}

func getSheetName(f *excel.File) (string, error) {
	sheets := f.GetSheetList()

	if len(sheets) == 0 {
		return "", fmt.Errorf("workbook contains no sheets")
	}

	if len(sheets) > 1 {
		return "", fmt.Errorf("expected one sheet, found %d", len(sheets))
	}

	return sheets[0], nil
}

func getHeaderIndexes(rows *excel.Rows, columns []string) ([]int, error) {
	headers, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Map column name -> index
	indexes := make(map[string]int)

	for i, header := range headers {
		indexes[header] = i
	}

	// Resolve requested columns
	columnsIndexes := make([]int, len(columns))

	for i, column := range columns {
		index, ok := indexes[column]
		if !ok {
			return nil, fmt.Errorf("column %q not found", column)
		}

		columnsIndexes[i] = index
	}

	return columnsIndexes, nil
}

func (e *ExcelFileManager) Read() error {
	if len(e.Columns) != 2 {
		return fmt.Errorf("columns must be 2 to perform read operations.")
	}

	f, err := excel.OpenFile(e.InputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	sheet, err := getSheetName(f)
	if err != nil {
		return err
	}

	// get dimensions to preallocate map
	dimension, err := f.GetSheetDimension(sheet)
	parts := strings.Split(dimension, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid dimension: %s", dimension)
	}

	_, endRow, err := excel.CellNameToCoordinates(parts[1])
	if err != nil {
		return err
	}

	e.Data = make(map[internal.IVID]internal.Product, endRow)

	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Read header
	if !rows.Next() {
		return fmt.Errorf("empty sheet")
	}

	columnsIndexes, err := getHeaderIndexes(rows, e.Columns)
	if err != nil {
		return err
	}

	// Read only requested fields
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return err
		}

		if len(row) <= columnsIndexes[1] {
			return fmt.Errorf("malformed row")
		}

		ivIDStr := row[columnsIndexes[0]]
		valueStr := row[columnsIndexes[1]]

		iv_id, err := strconv.Atoi(ivIDStr)

		if err != nil {
			return fmt.Errorf("Error: couldn't convert iv id: %v", ivIDStr)
		}

		product := internal.Product{
			IVID: internal.IVID(iv_id),
		}

		second_column := e.Columns[1]

		if second_column == "VAT" {
			floatVal, err := strconv.ParseFloat(valueStr, 32)
			if err != nil {
				return fmt.Errorf("Error: couldn't convert vat: %v", valueStr)
			}
			product.VAT = float32(floatVal)
		}

		if second_column == "SalePrice" {
			valueStr = strings.ReplaceAll(valueStr, ",", "")
			float_price, err := strconv.ParseFloat(valueStr, 32)
			if err != nil {
				return fmt.Errorf("Error: couldn't convert sale price: %v", valueStr)
			}
			product.Price = float32(float_price)
		}

		// check if the product already exists, if not, add it to the data map
		if foundProduct, ok := e.Data[internal.IVID(iv_id)]; !ok {
			e.Data[internal.IVID(iv_id)] = product
		} else {
			// if the found product's price is higher than the current product's price, skip don't update.
			// we want to keep the higher price
			if foundProduct.Price >= product.Price {
				continue
			}

			e.Data[internal.IVID(iv_id)] = product
		}
	}

	return rows.Error()
}

func (e *ExcelFileManager) Write() error {
	f := excel.NewFile()
	defer f.Close()

	err := f.SetCellValue("Sheet1", "A1", "iv_id")
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "B1", "vat")
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "C1", "price_before_vat")
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "D1", "price_after_vat")
	if err != nil {
		return err
	}

	row := 1
	for iv_id, product := range e.Data {
		row += 1
		err = f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), int(iv_id))
		if err != nil {
			return err
		}
		err = f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), product.VAT)
		if err != nil {
			return err
		}
		err = f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), product.Price)
		if err != nil {
			return err
		}
		err = f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), product.PriceAfterVAT)
		if err != nil {
			return err
		}
	}

	err = f.SaveAs(e.OutputFile)

	return err
}

func (e *ExcelFileManager) WriteStream() error {
	f := excel.NewFile()
	defer f.Close()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	// Header
	if err := sw.SetRow("A1", []interface{}{
		"iv_id",
		"vat",
		"price_before_vat",
		"price_after_vat",
	}); err != nil {
		return err
	}

	row := 2

	for ivID, product := range e.Data {
		cell := fmt.Sprintf("A%d", row)

		if err := sw.SetRow(cell, []interface{}{
			int(ivID),
			product.VAT,
			product.Price,
			product.PriceAfterVAT,
		}); err != nil {
			return err
		}

		row++
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	return f.SaveAs(e.OutputFile)
}

// used in reporting missing IVIDs in IMF and the ivids that are in IMF but not in the stock.
func (e *ExcelFileManager) WriteStreamReport(iv_ids []internal.IVID) error {
	f := excel.NewFile()
	defer f.Close()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	// Header
	if err := sw.SetRow("A1", []interface{}{
		"iv_id",
	}); err != nil {
		return err
	}

	row := 2
	for _, ivID := range iv_ids {
		cell := fmt.Sprintf("A%d", row)

		if err := sw.SetRow(cell, []interface{}{
			int(ivID),
		}); err != nil {
			return err
		}

		row++
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	return f.SaveAs(e.OutputFile)
}
