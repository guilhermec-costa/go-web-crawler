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

func (c *Controllers) HealthController(r *http.Request) (ControllerResponse, int) {
	return ControllerResponse{
		Data: "ok",
	}, http.StatusAccepted
}

func (c *Controllers) SignInController(r *http.Request) (ControllerResponse, int) {
	var payload services.SignInDTO
	if err := decodeJSON(r, &payload); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusBadRequest
	}

	if err := payload.Validate(); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity
	}

	token, err := services.SignInHandler(payload, c.app.UserStore)

	if err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusUnauthorized
	}

	return ControllerResponse{
		Data: token,
	}, http.StatusOK
}

func (c *Controllers) SignUpControler(r *http.Request) (ControllerResponse, int) {
	var payload services.SignUpDTO

	if err := decodeJSON(r, &payload); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity
	}

	if err := payload.Validate(); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusBadRequest
	}

	id, err := services.SignUpHandler(payload, c.app.UserStore)
	if err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError
	}

	return ControllerResponse{
		Data: StrAnyMap{
			"userId": id,
		},
	}, http.StatusOK
}

func (c *Controllers) TriggerCrawlerJobController(r *http.Request) (ControllerResponse, int) {
	var payload validation.CrawlerFlagsJSON
	if err := decodeJSON(r, &payload); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusBadRequest
	}

	params := validation.FromJSONToCrawlerParams(payload)
	if err := params.Validate(); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity
	}

	err := c.app.EnqueueCrawlerJob(app.Job{
		Params: params,
	})

	if err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusTooManyRequests
	}

	return ControllerResponse{
		Data: StrAnyMap{
			"message": "job started",
		},
	}, http.StatusOK
}
