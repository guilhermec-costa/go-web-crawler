package services

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/infra"
	"guilhermec-costa/go-web-crawler/server/types"
	"io"
	"log/slog"
	"os"
	"time"
)

type CrawlerService struct {
	crawlerStore infra.CrawlerExtractionDAO
	userStore    infra.UserDAO
}

func (s *CrawlerService) CreateExtraction(userId string, extractions string) (int64, error) {
	id, err := s.crawlerStore.Create(userId, extractions)
	if err != nil {
		slog.Error("failed to create extraction", "err", err)
		return 0, err
	}
	return id, err
}

func (s *CrawlerService) TriggerCrawlerExtraction(userId string, params validation.CrawlerParams, queue chan<- types.Job) error {
	job := types.Job{
		Params: params,
		UserId: userId,
	}

	// extractionId, saveErr := s.crawlerStore.Create(userId, "")
	select {
	case queue <- job:
		slog.Info("extraction job queued")
		return nil

	default:
		slog.Error("extraction queue is full")
	}

	return nil
}

func (s *CrawlerService) PatchExtractionByFilepathAndId(extractionId string, extractionPath string) error {
	file, fileOpErr := os.Open(extractionPath)
	if fileOpErr != nil {
		return fmt.Errorf("error opening extraction file")
	}

	defer file.Close()

	extractions, readErr := io.ReadAll(file)

	if readErr != nil {
		return fmt.Errorf("failed to read extractions file")
	}

	now := time.Now()
	patchData := infra.CrawlerExtractionUpdateDTO{
		Extractions: extractions,
		FinishedAt:  &now,
	}

	_, saveErr := s.crawlerStore.UpdateById(extractionId, patchData)
	if saveErr != nil {
		return fmt.Errorf("failed to save extraction")
	}

	slog.Info("extraction saved", "extractionId", extractionId)
	return nil
}

func NewCrawlerService(crawlerStore infra.CrawlerExtractionDAO, userStore infra.UserDAO) *CrawlerService {
	return &CrawlerService{
		crawlerStore: crawlerStore,
		userStore:    userStore,
	}
}
