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
	TimeoutMs  int
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

func ParseCrawlerFlags() (CrawlerFlags, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: crawler [options]\n\n")
		flag.PrintDefaults()
	}

	urlFlag := flag.String("url", "", "url for crawling")
	depthFlag := flag.Int("depth", 1, "depth for crawling")
	workersFlag := flag.Int("workers", 3, "Workers number for concurrent crawling")
	verboseFlag := flag.Bool("v", false, "run crawler on verbose mode")
	tickUpdateMs := flag.Int("tickupdate", 2000, "seconds between ticks")
	timeoutMs := flag.Int("timeout", 30000, "seconds to timeout")

	defaultPath := fmt.Sprintf("extractions-%s.jsonl", time.Now().Format("2006-01-02_15-04-05"))
	outputPathFlag := flag.String("o", defaultPath, "output path")

	flag.Parse()

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
		TimeoutMs:  *timeoutMs,
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
