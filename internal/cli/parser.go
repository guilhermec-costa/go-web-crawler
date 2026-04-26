package cli

import (
	"flag"
	"fmt"
	"os"
)

type CrawlerFlags struct {
	url     string
	depth   int
	workers int
}

func (f CrawlerFlags) String() string {
	return fmt.Sprintf(`---CliFlags---
url: %v
depth: %v
workers: %v`, f.url, f.depth, f.workers)
}

func ParseCrawlerFlags() CrawlerFlags {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: crawler [options]\n\n")
		flag.PrintDefaults()
	}

	urlFlag := flag.String("url", "", "url for crawling")
	depthFlag := flag.Int("depth", 1, "depth for crawling")
	workersFlag := flag.Int("workers", 3, "Workers number for concurrent crawling")

	flag.Parse()

	if len(*urlFlag) == 0 {
		fmt.Fprintln(os.Stderr, "Error: -url is required")
		flag.Usage()
		os.Exit(1)
	}

	return CrawlerFlags{
		url:     *urlFlag,
		depth:   *depthFlag,
		workers: *workersFlag,
	}
}

// func ParseUrl(url string)
