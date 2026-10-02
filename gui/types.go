package gui

// ProcessInput holds the user-supplied parameters from the frontend form.
type ProcessInput struct {
	OrgName           string `json:"orgName"`
	OutputDir         string `json:"outputDir"`
	IMFFilePath       string `json:"imfFilePath"`
	StockFilePath     string `json:"stockFilePath"`
	StockIVIDColumn   string `json:"stockIvidColumn"`
	StockPriceColumn  string `json:"stockPriceColumn"`
	IMFIVIDColumn     string `json:"imfIvidColumn"`
	IMFVATColumn      string `json:"imfVatColumn"`
}

// ProcessingResult is the DTO returned to the frontend after a successful run.
// Only summary statistics and generated filenames are included — never raw product data.
type ProcessingResult struct {
	ProductsProcessed int      `json:"productsProcessed"`
	StockProducts     int      `json:"stockProducts"`
	IMFProducts       int      `json:"imfProducts"`
	MissingInIMF      int      `json:"missingInIMF"`
	MissingInStock    int      `json:"missingInStock"`
	OutputFiles       []string `json:"outputFiles"` // basenames of generated files
	DurationMs        int64    `json:"durationMs"`
	OutputDir         string   `json:"outputDir"` // for "Open Folder" button
	Error             string   `json:"error,omitempty"`
}
