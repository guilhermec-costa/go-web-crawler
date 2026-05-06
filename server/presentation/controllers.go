package presentation

import (
	"fmt"
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

func (c *Controllers) SignInController(r *http.Request) (any, int) {
	var payload services.SignInDTO
	if err := decodeJSON(r, &payload); err != nil {
		return StrMap{"message": err.Error()}, http.StatusBadRequest
	}

	if err := payload.Validate(); err != nil {
		return StrMap{"message": err.Error()}, http.StatusUnprocessableEntity
	}

	token, err := services.SignInHandler(payload, c.app.UserStore)

	if err != nil {
		return ToMessageResponse(err.Error(), http.StatusUnprocessableEntity)
	}

	return ToMessageResponse(StrMap{"token": token}, http.StatusOK)
}

func (c *Controllers) SignUpControler(r *http.Request) (any, int) {
	var payload services.SignUpDTO

	if err := decodeJSON(r, &payload); err != nil {
		return ToMessageResponse(err.Error(), http.StatusUnprocessableEntity)
	}

	if err := payload.Validate(); err != nil {
		return ToMessageResponse(err.Error(), http.StatusBadRequest)
	}

	id, err := services.SignUpHandler(payload, c.app.UserStore)
	if err != nil {
		return ToMessageResponse(err.Error(), http.StatusInternalServerError)
	}

	return ToMessageResponse(fmt.Sprintf("id: %d", id), http.StatusOK)
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

	err := c.app.EnqueueCrawlerJob(app.Job{
		Params: params,
	})

	if err != nil {
		return StrMap{"message": err.Error()}, http.StatusTooManyRequests
	}

	return StrMap{"message": "job started"}, http.StatusOK
}
