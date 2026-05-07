package server

import (
	"guilhermec-costa/go-web-crawler/server/app"
	pres "guilhermec-costa/go-web-crawler/server/presentation"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func crawlerRouter(c *pres.Controllers) http.Handler {
	r := chi.NewRouter()
	r.Use(pres.AuthMiddleware)
	r.Post("/", pres.JsonHandler(c.TriggerCrawlerJobController))
	return r
}

func authRouter(c *pres.Controllers) http.Handler {
	r := chi.NewRouter()
	r.Post("/signin", pres.JsonHandler(c.SignInController))
	r.Post("/signup", pres.JsonHandler(c.SignUpControler))
	return r
}

func MakeCtrls(rootRouter *chi.Mux, app *app.App) *pres.Controllers {
	c := pres.NewControllers(app)
	rootRouter.Get("/health", pres.JsonHandler(c.HealthController))
	rootRouter.Mount("/auth", authRouter(c))
	rootRouter.Mount("/crawls", crawlerRouter(c))
	return c
}
