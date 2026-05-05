package presentation

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
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

	services.LoginHandler(payload)
	return StrMap{"message": "user logged"}, http.StatusOK
}

func (c *Controllers) TriggerCrawlerJobController(r *http.Request) (any, int) {
	var payload validation.CrawlerFlagsJSON
	if err := decodeJSON(r, &payload); err != nil {
		return StrMap{"message": err.Error()}, http.StatusBadRequest
	}

	params := validation.FromJSONToCrawlerParams(payload)
	if err := params.Validate(); err != nil {
		return StrMap{"message": err.Error()}, http.StatusUnprocessableEntity
	}

	c.app.EnqueueCrawlerJob(app.Job{
		Params: params,
	})

	return StrMap{"message": "job started"}, http.StatusOK
}
