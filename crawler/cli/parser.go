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

func (data *CrawlerFlags) Validate() error {
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

type CrawlerFlagsJSON struct {
	RootUrl string `json:"url"`
	Depth   int    `json:"depth"`
	Workers int    `json:"workers"`
}

func MergeWithDefault(args CrawlerFlagsJSON) (CrawlerFlags, error) {
	builtArgs := []string{
		fmt.Sprintf("-url=%s", args.RootUrl),
		fmt.Sprintf("-depth=%d", args.Depth),
		fmt.Sprintf("-workers=%d", args.Workers),
	}

	parsedFlags, err := ParseCrawlerFlags(builtArgs)
	if err != nil {
		if errors.Is(err, UrlRequiredErr) {
			return CrawlerFlags{}, fmt.Errorf("Failed to parse url: %w", err)
		}
	}
	return parsedFlags, nil
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

var UrlRequiredErr = errors.New("-url is required")

func ParseCrawlerFlags(args []string) (CrawlerFlags, error) {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)

	flag.Usage = func() {
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
	outputPathFlag := fs.String("o", da.OutputPath, "output path")

	fs.Parse(args)

	parsedFlags := CrawlerFlags{
		RootUrl:      *urlFlag,
		Depth:        *depthFlag,
		Workers:      *workersFlag,
		Verbose:      *verboseFlag,
		OutputPath:   *outputPathFlag,
		TickUpdateMs: *tickUpdateMs,
		TimeoutMs:    *timeoutMs,
	}

	if err := parsedFlags.Validate(); err != nil {
		return CrawlerFlags{}, err
	}

	return parsedFlags, nil
}

func ShowCrawlerConfigs(flags CrawlerFlags) {
	slog.Info("crawler config",
		"url", flags.RootUrl,
		"depth", flags.Depth,
		"workers", flags.Workers,
		"verbose", flags.Verbose,
	)
}
