package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"guilhermec-costa/go-web-crawler/server"
	"guilhermec-costa/go-web-crawler/server/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	const listenPort int16 = 3333
	slog.Info(fmt.Sprintf("Starting crawler server at port %d", listenPort))

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
