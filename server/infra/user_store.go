package infra

import (
	"database/sql"
	"fmt"
	"errors"
	"log/slog"
)

type User struct {
	Id        string
	Email     string
	Password  string
	CreatedAt string
}

type UserDAO interface {
	Create(email string, password string) (int64, error)
	FindByEmail(email string) (User, error)
}

type UserSQLiteStore struct {
	db *sql.DB
}

func (d *UserSQLiteStore) FindByEmail(email string) (User, error) {
	row := d.db.QueryRow(`
		SELECT id, email, password, created_at FROM users u
		WHERE u.email = ?
	`, email)

	var user User
	if err := row.Scan(&user.Id, &user.Email, &user.Password, &user.CreatedAt); err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return user, fmt.Errorf("user not found")
		}
		return user, err
	}

	return user, nil
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
