package app

import (
	"database/sql"
	"fmt"
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/infra"
	"log/slog"
	"path/filepath"
	"runtime"
)

type Job struct {
	Params validation.CrawlerParams
}

type App struct {
	JobQueue  chan Job
	UserStore infra.UserDAO
}

func (a *App) startJobQueueMonitor() {
	go func() {
		for job := range a.JobQueue {
			err := crawler.Bootstrap(job.Params)
			if err != nil {
				slog.Error("Failed crawlwer for job", "job", job)
			}
		}
	}()
}

func (a *App) EnqueueCrawlerJob(job Job) error {
	select {
	case a.JobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(b))
}

func NewApp() (*App, error) {
	dbPath := filepath.Join(RootDir(), "crawler.db")
	q := make(chan Job, 100)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("failed to open sqlite connection",
			"path", dbPath,
			"err", err,
		)
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	slog.Info("driver connection established")

	if err := db.Ping(); err != nil {
		slog.Error("database ping failed",
			"path", dbPath,
			"err", err,
		)
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	slog.Info("database initialized successfully", "path", dbPath)

	userStore := infra.NewUserSQLiteStore(db)

	slog.Info("running database migrations")
	if err := userStore.Migrate(); err != nil {
		slog.Error("migration failed", "err", err)
		return nil, fmt.Errorf("failed to migrate user store: %w", err)
	}

	app := &App{
		JobQueue:  q,
		UserStore: userStore,
	}

	slog.Info("starting job queue monitor", "buffer_size", cap(q))
	app.startJobQueueMonitor()

	return app, nil
}
