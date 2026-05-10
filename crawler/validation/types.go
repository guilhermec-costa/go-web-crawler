package validation

type CrawlerParams struct {
	RootUrl          string
	Depth            int
	Workers          int
	Verbose          bool
	OutputPath       string
	TickUpdateMs     int
	TimeoutMs        int
}

type CrawlerFlagsJSON struct {
	RootUrl string `json:"url"`
	Depth   int    `json:"depth"`
	Workers int    `json:"workers"`
}
