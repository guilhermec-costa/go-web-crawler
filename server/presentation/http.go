package presentation

import (
	"encoding/json"
	"net/http"
)

type StrAnyMap map[string]any

type ControllerResponse struct {
	Data  any `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func ToMessageResponse(data any, status int) (StrAnyMap, int) {
	return StrAnyMap{"message": data}, status
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}
