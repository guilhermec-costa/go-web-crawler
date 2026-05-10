package presentation

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/app"
	"guilhermec-costa/go-web-crawler/server/services"
	"net/http"
)

type Controllers struct {
	DI *app.DIContainer
}

func NewControllers(c *app.DIContainer) *Controllers {
	return &Controllers{
		DI: c,
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

	token, err := c.DI.UserService.SignInHandler(payload)

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

	id, err := c.DI.UserService.SignUpHandler(payload)
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

func (c *Controllers) ListExtractions(r *http.Request) (ControllerResponse, int) {
	vals := r.URL.Query()
	listDTO := services.ListExtractionDTO{}
	listDTO.ResolvePagination(vals.Get("page"), vals.Get("limit"))

	extractions, err := c.DI.CrawlerService.ListExtractions(listDTO)
	if err != nil {
		return ControllerResponse{
			Error: "failed to list extractions",
		}, http.StatusInternalServerError
	}

	return ControllerResponse{
		Data: extractions,
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

	err := c.DI.JobMonitor.TriggerJob(userId, params)

	if err != nil {
		return ControllerResponse{
			Error: err.Error(),
		}, http.StatusTooManyRequests
	}

	return ControllerResponse{
		Data: "job started",
	}, http.StatusOK
}
