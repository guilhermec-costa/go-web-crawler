package services

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/server/infra"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type CrawlerService struct {
	crawlerStore infra.CrawlerExtractionDAO
	userStore    infra.UserDAO
}

type ListExtractionDTO struct {
	Page  int64
	Limit int64
}

func parseIntOrDefault(value string, theDefault int64) int64 {
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return theDefault
	}
	return parsedValue
}

func (p *ListExtractionDTO) ResolvePagination(page string, limit string) {
	p.Page = parseIntOrDefault(page, 1)
	p.Limit = parseIntOrDefault(limit, 10)
}

func (s *CrawlerService) ListExtractions(payload ListExtractionDTO) ([]infra.CrawlerExtraction, error) {
	offset := (payload.Page - 1) * payload.Limit

	extractions, err := s.crawlerStore.List(payload.Limit, offset)
	if err != nil {
		slog.Error("failed to list extractions", "err", err)
		return []infra.CrawlerExtraction{}, err
	}

	return extractions, nil
}

func (s *CrawlerService) CreateExtraction(userId string, extractions string) (int64, error) {
	id, err := s.crawlerStore.Create(userId, extractions)
	if err != nil {
		slog.Error("failed to create extraction", "err", err)
		return 0, err
	}
	return id, err
}

func (s *CrawlerService) PatchExtractionContentFromFilepath(extractionId string, extractionPath string) error {
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
