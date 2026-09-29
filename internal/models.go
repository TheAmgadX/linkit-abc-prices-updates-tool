package internal

type IVID int
type FileInput map[IVID]Product

type Product struct {
	IVID          IVID
	VAT           float32 // got from the IMF input file
	Price         float32 // got from the Stock input file
	PriceAfterVAT float32 // generated in the program
}
