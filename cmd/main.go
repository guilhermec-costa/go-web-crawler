package main

import (
	"guilhermec-costa/go-web-crawler/internal/cli"
	"guilhermec-costa/go-web-crawler/internal/crawler"
	"guilhermec-costa/go-web-crawler/internal/log"
)

func main() {
	log.SetupLogger()

	flags, err := cli.ParseCrawlerFlags()
	if err != nil {
		cli.ExitWithFlagUsage(err.Error())
	}

	cli.ShowCrawlerConfigs(flags)
	crawler.Bootstrap(flags)
}
