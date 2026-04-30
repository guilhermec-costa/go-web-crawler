package main

import (
	"guilhermec-costa/go-web-crawler/internal/cli"
	"guilhermec-costa/go-web-crawler/internal/crawler"
	"log/slog"
	"os"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	flags, err := cli.ParseCrawlerFlags()
	if err != nil {
		cli.ExitWithFlagUsage(err.Error())
	}

	cli.ShowCrawlerConfigs(flags)
	crawler.Bootstrap(flags)
}
