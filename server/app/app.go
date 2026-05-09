package app

import (
	"database/sql"
	"fmt"
	"guilhermec-costa/go-web-crawler/server/infra"
	"guilhermec-costa/go-web-crawler/server/services"
	"guilhermec-costa/go-web-crawler/server/types"
	"log/slog"
	"path/filepath"
	"runtime"
)

type App struct {
	JobQueue               chan types.Job
	UserStore              infra.UserDAO
	CrawlerExtractionStore infra.CrawlerExtractionDAO
	JobProcessor           types.JobProcessor
	UserService            *services.AuthService
	CrawlerService         *services.CrawlerService
}

func (a *App) startCrawlerJobQueueMonitor() {
	go func() {
		for job := range a.JobQueue {
			if err := a.JobProcessor(job); err != nil {
				slog.Error("failed to process job", "err", err)
			}
		}
	}()
}

func (a *App) SetJobProcessor(jp types.JobProcessor) {
	a.JobProcessor = jp
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(b))
}

func NewApp() (*App, error) {
	dbPath := filepath.Join(RootDir(), "crawler.db")
	q := make(chan types.Job, 100)

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
	crawlerExtractionStore := infra.NewCrawlerSQLiteStore(db)

	slog.Info("running database migrations")
	if err := userStore.Migrate(); err != nil {
		slog.Error("user migration failed", "err", err)
		return nil, fmt.Errorf("failed to migrate user store: %w", err)
	}

	if err := crawlerExtractionStore.Migrate(); err != nil {
		slog.Error("crawler extraction migration failed", "err", err)
		return nil, fmt.Errorf("failed to migrate crawler extraction store: %w", err)
	}

	app := &App{
		JobQueue:               q,
		UserStore:              userStore,
		CrawlerExtractionStore: crawlerExtractionStore,
		UserService:            services.NewAuthService(userStore),
		CrawlerService:         services.NewCrawlerService(crawlerExtractionStore, userStore),
	}

	slog.Info("starting job queue monitor", "buffer_size", cap(q))

	app.startCrawlerJobQueueMonitor()

	return app, nil
}
