package main

import (
	"errors"
	"log/slog"
	_ "modernc.org/sqlite"
	"net/http"
	"os"

	"guilhermec-costa/go-web-crawler/server"
	"guilhermec-costa/go-web-crawler/server/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting crawler server at port 3333")
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	app, err := app.NewApp()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	server.SetupServerRoutes(r, app)

	if err := http.ListenAndServe(":3333", r); err != nil {
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
