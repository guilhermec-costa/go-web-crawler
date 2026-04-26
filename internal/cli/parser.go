package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type CrawlerFlags struct {
	Url     string
	Depth   int
	Workers int
}

func (f CrawlerFlags) String() string {
	return fmt.Sprintf(
		"URL     : %s\nDepth   : %d\nWorkers : %d",
		f.Url, f.Depth, f.Workers,
	)
}

func ExitWithFlagUsage(message string) {
	fmt.Fprintln(os.Stderr, )
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
		Url:     *urlFlag,
		Depth:   *depthFlag,
		Workers: *workersFlag,
	}, nil
}

func ShowCrawlerConfigs(flags CrawlerFlags) {
	fmt.Println("Crawler configuration")
	fmt.Println("------------------------")
	fmt.Println(flags)
	fmt.Println()
}
