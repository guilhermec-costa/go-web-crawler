package cli

import (
	"errors"
	"flag"
	"fmt"
	val "guilhermec-costa/go-web-crawler/crawler/validation"
	"log/slog"
	"os"
)

func ExitWithFlagUsage(message string) {
	fmt.Fprintln(os.Stderr)
	flag.Usage()
	os.Exit(1)
}

var UrlRequiredErr = errors.New("-url is required")

func ParseCliCrawlerFlags(args []string) (val.CrawlerParams, error) {
	fs := flag.NewFlagSet("crawler", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: crawler [options]\n\n")
		fs.PrintDefaults()
	}

	da := val.DefaultCrawlerParams()
	urlFlag := fs.String("url", da.RootUrl, "url for crawling")
	depthFlag := fs.Int("depth", da.Depth, "depth for crawling")
	workersFlag := fs.Int("workers", da.Workers, "Workers number for concurrent crawling")
	verboseFlag := fs.Bool("v", da.Verbose, "run crawler on verbose mode")
	tickUpdateMs := fs.Int("tickupdate", da.TickUpdateMs, "seconds between ticks")
	timeoutMs := fs.Int("timeout", da.TimeoutMs, "seconds to timeout")
	outputPathFlag := fs.String("o", da.OutputPath, "output path")

	fs.Parse(args)

	if len(*urlFlag) == 0 {
		return val.CrawlerParams{}, UrlRequiredErr
	}

	parsedFlags := val.CrawlerParams{
		RootUrl:      *urlFlag,
		Depth:        *depthFlag,
		Workers:      *workersFlag,
		Verbose:      *verboseFlag,
		OutputPath:   *outputPathFlag,
		TickUpdateMs: *tickUpdateMs,
		TimeoutMs:    *timeoutMs,
	}

	if err := parsedFlags.Validate(); err != nil {
		return val.CrawlerParams{}, err
	}

	return parsedFlags, nil
}

func ShowCrawlerConfigs(flags val.CrawlerParams) {
	slog.Info("crawler config",
		"url", flags.RootUrl,
		"depth", flags.Depth,
		"workers", flags.Workers,
		"verbose", flags.Verbose,
	)
}
