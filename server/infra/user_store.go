package infra

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type User struct {
	email string
}

type UserDAO interface {
	Create(email string, password string) (int64, error)
}

type UserSQLiteStore struct {
	db *sql.DB
}

func (d *UserSQLiteStore) Create(email string, password string) (int64, error) {
	result, err := d.db.Exec(`
		INSERT INTO users (email, password) VALUES (?, ?)
		RETURNING id
	`, email, password)

	if err != nil {
		slog.Error("Failed to insert user in database", "err", err)
		return 0, fmt.Errorf("failed inserting user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed getting last inserted id: %v", err)
	}

	return id, nil
}

func (d *UserSQLiteStore) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);	
	`)

	return err
}

func NewUserSQLiteStore(db *sql.DB) *UserSQLiteStore {
	return &UserSQLiteStore{
		db: db,
	}
}
