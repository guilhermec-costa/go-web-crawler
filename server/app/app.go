package app

import (
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/cli"
)

type Job struct {
	id   string
	args cli.CrawlerFlags
}

type App struct {
	JobQueue chan Job
}

func startJobQueueMonitor(queue <-chan Job) {
	for job := range queue {
		crawler.Bootstrap(job.args)
	}
}

func NewApp() *App {
	q := make(chan Job, 100)

	app := &App{
		JobQueue: q,
	}

	return app
}
