package infra

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type CrawlerExtraction struct {
	Id          string
	UserId      string
	Extractions []byte
	CreatedAt   *time.Time
	FinishedAt  *time.Time
}

type CrawlerExtractionUpdateDTO struct {
	Extractions []byte
	FinishedAt  *time.Time
}

type CrawlerExtractionDAO interface {
	Create(userId string, extractions string) (int64, error)
	UpdateById(extractionId string, patchData CrawlerExtractionUpdateDTO) (int64, error)
	FindById(id string) (CrawlerExtraction, error)
	List(page int, limit int) ([]CrawlerExtraction, error)
}

type CrawlerExtractionSQLiteStore struct {
	db *sql.DB
}

func (d *CrawlerExtractionSQLiteStore) findByColumn(columnName string, value any) (CrawlerExtraction, error) {
	queryStr := fmt.Sprintf(`
		SELECT id, user_id, extraction, created_at, finished_at FROM crawler_jobs j
		where j.%s = ?
	`, columnName)

	row := d.db.QueryRow(queryStr, value)

	var extraction CrawlerExtraction
	var finishedAt sql.NullTime

	err := row.Scan(
		&extraction.Id,
		&extraction.UserId,
		&extraction.Extractions,
		&extraction.CreatedAt,
		&finishedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return extraction, fmt.Errorf("extraction not found by %s", columnName)
		}
		return extraction, err
	}

	if finishedAt.Valid {
		extraction.FinishedAt = &finishedAt.Time
	}
	return extraction, nil
}

func (d *CrawlerExtractionSQLiteStore) FindById(id string) (CrawlerExtraction, error) {
	return d.findByColumn("id", id)
}

func (s *CrawlerExtractionSQLiteStore) List(page int, limit int) ([]CrawlerExtraction, error) {
	// var extractions []CrawlerExtraction
	return []CrawlerExtraction{}, nil
}

func (s *CrawlerExtractionSQLiteStore) Create(userId string, extractions string) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO crawler_jobs (user_id, extraction)
		VALUES (?, ?);
	`, userId, extractions)

	if err != nil {
		slog.Error("failed to create extraction register", "err", err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed getting last inserted id: %v", err)
	}

	return id, nil
}

func (s *CrawlerExtractionSQLiteStore) UpdateById(extractionId string, payload CrawlerExtractionUpdateDTO) (int64, error) {
	cur, err := s.FindById(extractionId)
	if err != nil {
		return 0, err
	}

	if payload.Extractions != nil {
		cur.Extractions = payload.Extractions
	}

	if payload.FinishedAt != nil {
		cur.FinishedAt = payload.FinishedAt
	}

	result, err := s.db.Exec(`
		UPDATE crawler_jobs
		SET extraction = ?, finished_at = ?
		WHERE id = ?
	`, cur.Extractions, cur.FinishedAt, cur.Id)

	if err != nil {
		slog.Error("failed to update extraction row")
		return 0, err
	}

	affec, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affec, nil
}

func NewCrawlerSQLiteStore(db *sql.DB) *CrawlerExtractionSQLiteStore {
	return &CrawlerExtractionSQLiteStore{
		db: db,
	}
}
