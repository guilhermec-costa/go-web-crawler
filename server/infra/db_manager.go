package infra

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type DbManager struct {
	db      *sql.DB
	connStr string
}

func (d *DbManager) DB() *sql.DB {
	return d.db;
}
func NewDbManager(driver string, connStr string) (*DbManager, error) {
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		slog.Error("failed to open sqlite connection",
			"connStr", connStr,
			"err", err,
		)
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	slog.Info("driver connection established")

	if err := db.Ping(); err != nil {
		slog.Error("database ping failed",
			"connStr", connStr,
			"err", err,
		)
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	slog.Info("database initialized successfully", "connStr", connStr)
	return &DbManager{
		db:      db,
		connStr: connStr,
	}, nil
}

func (d *DbManager) Ping() error {
	if err := d.db.Ping(); err != nil {
		slog.Error("database ping failed",
			"connStr", d.connStr,
			"err", err,
		)
		return fmt.Errorf("database unreachable: %w", err)
	}
	return nil
}
