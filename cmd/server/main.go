package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	_ "modernc.org/sqlite"

	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/server"
	"guilhermec-costa/go-web-crawler/server/app"
	"guilhermec-costa/go-web-crawler/server/infra"
	"guilhermec-costa/go-web-crawler/server/services"
	"guilhermec-costa/go-web-crawler/server/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(b))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	const listenPort int16 = 3333
	slog.Info(fmt.Sprintf("Starting crawler server at port %d", listenPort))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	dbPath := filepath.Join(RootDir(), "crawler.db")
	dbMng, err := infra.NewDbManager("sqlite", dbPath)
	if err != nil {
		slog.Error("database connection failed: %v", "err", err)
		return
	}

	slog.Info("driver connection established")

	if err := dbMng.Ping(); err != nil {
		slog.Error("database unreachable: %v", "err", err)
		return
	}

	slog.Info("database initialized successfully", "path", dbPath)

	userStore := infra.NewUserSQLiteStore(dbMng.DB())
	crawlerExtractionStore := infra.NewCrawlerSQLiteStore(dbMng.DB())

	migrators := []infra.Migrator{userStore, crawlerExtractionStore}
	for _, m := range migrators {
		if err := m.Migrate(); err != nil {
			slog.Error("migration failed", "err", err)
			return
		}
	}

	userService := services.NewAuthService(userStore)
	crawlerService := services.NewCrawlerService(crawlerExtractionStore, userStore)
	jobMonitor := infra.NewJobMonitor(func(job types.Job) error {
		extractionId, createErr := crawlerService.CreateExtraction(job.UserId, "[]")
		if createErr != nil {
			return createErr
		}

		outputPath, bstrapErr := crawler.Bootstrap(job.Params)
		if bstrapErr != nil {
			return bstrapErr
		}

		extrErr := crawlerService.PatchExtractionContentFromFilepath(strconv.FormatInt(extractionId, 10), outputPath)
		if extrErr != nil {
			return extrErr
		}
		if remErr := os.Remove(outputPath); remErr != nil {
			slog.Error("failed to remove output file", "err", remErr)
			return extrErr
		}
		slog.Info("removed output file", "otpPath", outputPath)
		return nil
	}, 100)

	diContainer := &app.DIContainer{
		JobMonitor:             jobMonitor,
		UserStore:              userStore,
		CrawlerExtractionStore: crawlerExtractionStore,
		UserService:            userService,
		CrawlerService:         crawlerService,
	}

	server.MakeCtrls(r, diContainer)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), r); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			{
				slog.Error("Server closed")
			}

		default:
			slog.Error("Error starting server", "err", err)
		}
	}
}
