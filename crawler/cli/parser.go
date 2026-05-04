package cli

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type CrawlerFlags struct {
	RootUrl      string
	Depth        int
	Workers      int
	Verbose      bool
	OutputPath   string
	TickUpdateMs int
	TimeoutMs    int
}

type CrawlerFlagsJSON struct {
	RootUrl string `json:"url"`
	Depth   int    `json:"depth"`
}

func (data *CrawlerFlagsJSON) Validate() error {
	if len(data.RootUrl) == 0 {
		return fmt.Errorf("url is required")
	}

	if data.Depth < 0 {
		return fmt.Errorf("depth must be greater than  0")
	}
	return nil
}

func (f CrawlerFlags) String() string {
	return fmt.Sprintf(
		"Root URL     : %s\nDepth   : %d\nWorkers : %d",
		f.RootUrl, f.Depth, f.Workers,
	)
}

func ExitWithFlagUsage(message string) {
	fmt.Fprintln(os.Stderr)
	flag.Usage()
	os.Exit(1)
}

func DefaultArgs() CrawlerFlags {
	defaultPath := fmt.Sprintf("extractions-%s.jsonl", time.Now().Format("2006-01-02_15-04-05"))
	return CrawlerFlags{
		RootUrl:      "",
		Depth:        1,
		Workers:      3,
		Verbose:      false,
		TickUpdateMs: 2000,
		OutputPath:   defaultPath,
		TimeoutMs:    30000,
	}
}

func ParseCrawlerFlags(args []string) (CrawlerFlags, error) {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: crawler [options]\n\n")
		fs.PrintDefaults()
	}

	da := DefaultArgs()
	urlFlag := fs.String("url", da.RootUrl, "url for crawling")
	depthFlag := fs.Int("depth", da.Depth, "depth for crawling")
	workersFlag := fs.Int("workers", da.Workers, "Workers number for concurrent crawling")
	verboseFlag := fs.Bool("v", da.Verbose, "run crawler on verbose mode")
	tickUpdateMs := fs.Int("tickupdate", da.TickUpdateMs, "seconds between ticks")
	timeoutMs := fs.Int("timeout", da.TimeoutMs, "seconds to timeout")

	defaultPath := fmt.Sprintf("extractions-%s.jsonl", time.Now().Format("2006-01-02_15-04-05"))
	outputPathFlag := fs.String("o", defaultPath, "output path")

	fs.Parse(args)

	if len(*urlFlag) == 0 {
		return CrawlerFlags{}, errors.New("-url is required")
	}

	return CrawlerFlags{
		RootUrl:      *urlFlag,
		Depth:        *depthFlag,
		Workers:      *workersFlag,
		Verbose:      *verboseFlag,
		OutputPath:   *outputPathFlag,
		TickUpdateMs: *tickUpdateMs,
		TimeoutMs:    *timeoutMs,
	}, nil
}

func ShowCrawlerConfigs(flags CrawlerFlags) {
	slog.Info("crawler config",
		"url", flags.RootUrl,
		"depth", flags.Depth,
		"workers", flags.Workers,
		"verbose", flags.Verbose,
	)
}
