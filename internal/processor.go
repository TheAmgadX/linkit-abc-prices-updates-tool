package internal

type Processor struct {
	StockData *map[IVID]Product
	IMFData   *map[IVID]Product

	NotFoundIVID       []IVID // IVIDs in Stock data not found in IMF data
	HaveNoMatchingIVID []IVID // IVIDs in IMF data not found in Stock data
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
			p.NotFoundIVID = append(p.NotFoundIVID, ivid)
		}
	}

	for ivid := range *p.IMFData {
		if _, ok := result[ivid]; !ok {
			p.HaveNoMatchingIVID = append(p.HaveNoMatchingIVID, ivid)
		}
	}

	return result
}
