package crawler

import (
	"guilhermec-costa/go-web-crawler/internal/cli"
	"log"
	"net/url"
)

func StartCrawlerEngine(args cli.CrawlerFlags) {
	log.Printf("Starting crawler for: %s", args.Url)

	_url, err := url.Parse(args.Url)
	if err != nil {
		log.Fatalf("[ERROR] url %s is not valid: %v", args.Url, err)
	}

	if err := ValidateUrl(_url); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	log.Println(_url)
}
