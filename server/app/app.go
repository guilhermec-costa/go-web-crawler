package app

import (
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"log/slog"
)

type Job struct {
	Params validation.CrawlerParams
}

type App struct {
	JobQueue chan Job
}

func (a *App) startJobQueueMonitor() {
	go func() {
		for job := range a.JobQueue {
			err := crawler.Bootstrap(job.Params)
			if err != nil {
				slog.Error("Failed crawlwer for job", "job", job)
			}
		}
	}()
}

func (a *App) EnqueueCrawlerJob(job Job) {
	go func() {
		slog.Info("Enqueuing crawler job", "params", job.Params)
		a.JobQueue <- job
		slog.Info("Job enqueued")
	}()
}

func NewApp() *App {
	q := make(chan Job, 100)

	app := &App{
		JobQueue: q,
	}

	app.startJobQueueMonitor()

	return app
}
