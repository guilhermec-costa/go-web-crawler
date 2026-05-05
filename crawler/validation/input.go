package validation

import (
	"fmt"
	"log/slog"
	"time"
)

func (data *CrawlerParams) Validate() error {
	if len(data.RootUrl) == 0 {
		slog.Error("url is required")
		return fmt.Errorf("url is required")
	}

	if data.Depth < 0 {
		slog.Error("depth must be greater than 0")
		return fmt.Errorf("depth must be greater than  0")
	}

	if data.Workers <= 0 || data.Workers > 10 {
		slog.Error("maximum of 10 workers allowed")
		return fmt.Errorf("maximum of 10 workers allowed")
	}

	return nil
}

func DefaultCrawlerParams() CrawlerParams {
	defaultPath := fmt.Sprintf("extractions-%s.jsonl", time.Now().Format("2006-01-02_15-04-05"))
	return CrawlerParams{
		RootUrl:      "",
		Depth:        1,
		Workers:      3,
		Verbose:      false,
		TickUpdateMs: 2000,
		OutputPath:   defaultPath,
		TimeoutMs:    30000,
	}
}

func FromJSONToCrawlerParams(payload CrawlerFlagsJSON) CrawlerParams {
	defaults := DefaultCrawlerParams()
	params := CrawlerParams{
		RootUrl:      payload.RootUrl,
		Depth:        payload.Depth,
		Workers:      payload.Workers,
		Verbose:      defaults.Verbose,
		OutputPath:   defaults.OutputPath,
		TickUpdateMs: defaults.TickUpdateMs,
		TimeoutMs:    defaults.TimeoutMs,
	}

	return params
}

func (f CrawlerParams) String() string {
	return fmt.Sprintf(
		"Root URL     : %s\nDepth   : %d\nWorkers : %d",
		f.RootUrl, f.Depth, f.Workers,
	)
}
