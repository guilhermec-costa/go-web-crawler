package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type CrawlerFlags struct {
	RootUrl     string
	Depth   int
	Workers int
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

	flag.Parse()

	if len(*urlFlag) == 0 {
		return CrawlerFlags{}, errors.New("-url is required")
	}

	return CrawlerFlags{
		RootUrl:     *urlFlag,
		Depth:   *depthFlag,
		Workers: *workersFlag,
	}, nil
}

func ShowCrawlerConfigs(flags CrawlerFlags) {
	log.Printf(
		"Using URL: %s | Depth: %d | Workers: %d\n",
		flags.RootUrl,
		flags.Depth,
		flags.Workers,
	)
}
