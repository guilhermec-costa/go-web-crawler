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

func NewUserSQLiteStore(db *sql.DB) *UserSQLiteStore {
	return &UserSQLiteStore{
		db: db,
	}
}
