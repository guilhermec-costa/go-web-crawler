package presentation

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/app"
	"guilhermec-costa/go-web-crawler/server/services"
	"net/http"
)

type Controllers struct {
	a *app.App
}

func NewControllers(a *app.App) *Controllers {
	return &Controllers{
		a: a,
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

	token, err := c.a.UserService.SignInHandler(payload)

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
		}, http.StatusBadRequest
	}

	if err := payload.Validate(); err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity
	}

	id, err := c.a.UserService.SignUpHandler(payload)
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

	userId, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		return ControllerResponse{
			Error: "unauthorized",
		}, http.StatusUnauthorized
	}

	err := c.a.CrawlerService.TriggerCrawlerExtraction(userId, params, c.a.JobQueue)

	if err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusTooManyRequests
	}

	return ControllerResponse{
		Data: "job started",
	}, http.StatusOK
}
