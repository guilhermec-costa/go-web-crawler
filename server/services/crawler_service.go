package services

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/infra"
	"guilhermec-costa/go-web-crawler/server/types"
	"log/slog"
	// "os"
)

type CrawlerService struct {
	store infra.CrawlerExtractionDAO
}

func (s *CrawlerService) TriggerCrawlerExtraction(userId string, params validation.CrawlerParams, queue chan<- types.Job) error {
	job := types.Job{
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

func (s *CrawlerService) SaveCrawlerExtraction(userId string, extractionPath string, store infra.CrawlerExtractionDAO) error {
	// store.Create()
	// s.store.Create(userId, )
	slog.Info("user that triggered", "user", userId)
	return nil
}

func NewCrawlerService(s infra.CrawlerExtractionDAO) *CrawlerService {
	return &CrawlerService{
		store: s,
	}
}
