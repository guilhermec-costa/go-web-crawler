package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/server"
	"guilhermec-costa/go-web-crawler/server/app"
	"guilhermec-costa/go-web-crawler/server/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func checkEnv() {
	if err := godotenv.Load(); err != nil {
		panic("failed to load env variables")
	}

	if _, found := os.LookupEnv("JWTSECRET"); !found {
		panic("jwt secret not found. Verify if is JWTSECRET is set")
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// checkEnv()

	const listenPort int16 = 3333
	slog.Info(fmt.Sprintf("Starting crawler server at port %d", listenPort))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	theApp, err := app.NewApp()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	theApp.SetJobProcessor(func(job types.Job) error {
		outputPath, bstrapErr := crawler.Bootstrap(job.Params)
		if bstrapErr != nil {
			return bstrapErr
		}
		extrErr := theApp.CrawlerService.SaveCrawlerExtraction(job.UserId, outputPath, theApp.CrawlerExtractionStore)
		if extrErr != nil {
			return extrErr
		}
		return nil
	})

	server.MakeCtrls(r, theApp)

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
