package infra

import (
	"database/sql"
)

type CrawlerExtraction struct {
	Id          string
	Extractions string
}

type CrawlerExtractionUpdate struct {
	UpdatedAt string
}

type CrawlerExtractionDAO interface {
	Create(userId string, extractions string) (int, error)
	Update(CrawlerExtractionUpdate)
}

type CrawlerExtractionSQLiteStore struct {
	db *sql.DB
}

func (s *CrawlerExtractionSQLiteStore) Create(userId string, extractions string) (int, error) {
	_, err := s.db.Exec(`
		INSERT INTO crawljobs ()
	`)

	if err != nil {
		return 0, err
	}

	return 0, nil
}

func (s *CrawlerExtractionSQLiteStore) Update(payload CrawlerExtractionUpdate) {

}

func NewCrawlerSQLiteStore(db *sql.DB) *CrawlerExtractionSQLiteStore {
	return &CrawlerExtractionSQLiteStore{
		db: db,
	}
}
