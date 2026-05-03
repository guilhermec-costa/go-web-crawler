package main

import (
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/cli"
	"log/slog"
	"os"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	flags, err := cli.ParseCrawlerFlags(os.Args[1:])
	if err != nil {
		cli.ExitWithFlagUsage(err.Error())
	}

	cli.ShowCrawlerConfigs(flags)
	crawler.Bootstrap(flags)
}
