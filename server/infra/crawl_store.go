package infra

import (
	"database/sql"
	"fmt"
)

type CrawlerExtraction struct {
	Id          string
	Extractions string
}

type CrawlerExtractionUpdate struct {
	UpdatedAt string
}

type CrawlerExtractionDAO interface {
	Create(userId string, extractions string) (int64, error)
	Update(CrawlerExtractionUpdate)
}

type CrawlerExtractionSQLiteStore struct {
	db *sql.DB
}

func (s *CrawlerExtractionSQLiteStore) Create(userId string, extractions string) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO crawljobs (user_id, extractions)
		VALUES (?, ?);
	`, userId, extractions)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed getting last inserted id: %v", err)
	}

	return id, nil
}

func (s *CrawlerExtractionSQLiteStore) Update(payload CrawlerExtractionUpdate) {
}

func NewCrawlerSQLiteStore(db *sql.DB) *CrawlerExtractionSQLiteStore {
	return &CrawlerExtractionSQLiteStore{
		db: db,
	}
}
