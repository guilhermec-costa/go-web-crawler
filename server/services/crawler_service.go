package services

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/app"
	"log/slog"
	// "os"
)

func TriggerCrawlerExtraction(userId string, params validation.CrawlerParams, queue chan<- app.Job) error {
	job := app.Job{
		Params: params,
		UserId: userId,
	}

	select {
	case queue <- job:
		slog.Info("extraction job queued")
		return nil

	default:
		slog.Error("extraction queue is full")
	}

	return nil
}

func SaveCrawlerExtraction(extractionPath string) error {
	return nil
}
