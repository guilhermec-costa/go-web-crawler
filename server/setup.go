package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func crawlerRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/", startCrawlerJob)
	return r
}

func SetupServerRoutes(rootRouter *chi.Mux) {
	rootRouter.Get("/health", health)
	rootRouter.Mount("/crawls", crawlerRouter())
}
