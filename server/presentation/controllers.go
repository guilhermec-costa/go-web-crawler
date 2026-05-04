package presentation

import (
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/cli"
	"guilhermec-costa/go-web-crawler/server/app"
	"guilhermec-costa/go-web-crawler/server/services"
	"net/http"
)

type Controllers struct {
	app *app.App
}

func NewControllers(a *app.App) *Controllers {
	return &Controllers{
		app: a,
	}
}

func (c *Controllers) HealthController(r *http.Request) (any, int) {
	return StrMap{"status": "ok"}, http.StatusOK
}

func (c *Controllers) LoginController(r *http.Request) (any, int) {
	var payload services.LoginDTO
	if err := decodeJSON(r, &payload); err != nil {
		return StrMap{"message": err.Error()}, http.StatusBadRequest
	}

	if err := payload.Validate(); err != nil {
		return StrMap{"message": err.Error()}, http.StatusUnprocessableEntity
	}

	services.AuthService(payload)
	return StrMap{"message": "user loggedf"}, http.StatusOK
}

func (c *Controllers) TriggerCrawlerJobController(r *http.Request) (any, int) {
	var payload cli.CrawlerFlagsJSON
	if err := decodeJSON(r, &payload); err != nil {
		return StrMap{"message": err.Error()}, http.StatusBadRequest
	}

	if crawlerArgs, err := cli.MergeWithDefault(payload); err != nil {
		return StrMap{"message": "failed parsing crawler flags:" + err.Error()}, http.StatusBadRequest
	} else {
		go crawler.Bootstrap(crawlerArgs)
	}

	return StrMap{"message": "job started"}, http.StatusOK
}
