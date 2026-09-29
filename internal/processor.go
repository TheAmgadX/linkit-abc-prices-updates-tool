package internal

import "fmt"

type Processor struct {
	StockData *map[IVID]Product
	IMFData   *map[IVID]Product
}

func NewProcessor(stockData, imfData *map[IVID]Product) *Processor {
	return &Processor{
		StockData: stockData,
		IMFData:   imfData,
	}
}

func (p *Processor) Process() map[IVID]Product {
	result := make(map[IVID]Product, len(*p.StockData))
	for ivid, stock := range *p.StockData {
		if imf, ok := (*p.IMFData)[ivid]; ok {
			stock.VAT = imf.VAT
			stock.PriceAfterVAT = stock.Price * (1 + imf.VAT)
			result[ivid] = stock
		} else {
			fmt.Printf("Item: %v, Not found in IMF data\n", ivid)
		}
	}

	return result
}
