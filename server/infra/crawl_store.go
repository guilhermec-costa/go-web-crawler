package infra

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

func jsonlStrToJSONArray(jsonl string) (json.RawMessage, error) {
	lines := strings.Split(strings.TrimSpace(jsonl), "\n")

	var objects []json.RawMessage

	for _, line := range lines {
		objects = append(objects, json.RawMessage(line))
	}

	return json.Marshal(objects)
}

type CrawlerExtraction struct {
	Id          string
	UserId      string
	Extractions json.RawMessage `json:"extraction"`
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
	List(limit int64, offset int64) ([]CrawlerExtraction, error)
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
	var createdAt sql.NullTime
	var finishedAt sql.NullTime
	var extractionRawJSON string

	err := row.Scan(
		&extraction.Id,
		&extraction.UserId,
		&extractionRawJSON,
		&createdAt,
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

	if createdAt.Valid {
		extraction.CreatedAt = &createdAt.Time
	}

	if parsedExtJson, err := jsonlStrToJSONArray(extractionRawJSON); err != nil {
		slog.Error("failed to parse json", "err", err)
		extraction.Extractions = json.RawMessage("[]")
	} else {
		extraction.Extractions = parsedExtJson
	}

	return extraction, nil
}

func (d *CrawlerExtractionSQLiteStore) FindById(id string) (CrawlerExtraction, error) {
	return d.findByColumn("id", id)
}

func (s *CrawlerExtractionSQLiteStore) List(limit int64, offset int64) ([]CrawlerExtraction, error) {
	var extractions []CrawlerExtraction

	rows, err := s.db.Query(`
		SELECT id, user_id, extraction, created_at, finished_at FROM crawler_jobs
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return extractions, err
	}

	defer rows.Close()

	for rows.Next() {
		var extraction CrawlerExtraction
		var createdAt sql.NullTime
		var extractionRawJSON string
		var finishedAt sql.NullTime

		err := rows.Scan(
			&extraction.Id,
			&extraction.UserId,
			&extractionRawJSON,
			&createdAt,
			&finishedAt,
		)

		if err != nil {
			return extractions, err
		}

		if createdAt.Valid {
			extraction.CreatedAt = &createdAt.Time
		}

		if finishedAt.Valid {
			extraction.FinishedAt = &finishedAt.Time
		}

		if parsedExtJson, err := jsonlStrToJSONArray(extractionRawJSON); err != nil {
			slog.Error("failed to parse json", "err", err)
			extraction.Extractions = json.RawMessage("[]")
		} else {
			extraction.Extractions = parsedExtJson
		}
		extractions = append(extractions, extraction)
	}

	if err := rows.Err(); err != nil {
		slog.Error("error querying extractios", "limit", limit, "offset", offset, "err", err)
		return extractions, err
	}

	return extractions, nil
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
