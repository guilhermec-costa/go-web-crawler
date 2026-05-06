package presentation

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type StrMap map[string]string
type StrAnyMap map[string]any

func SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func ToJSONResponse(w io.Writer, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "err", err)
		return
	}

	fmt.Fprintf(w, "%v", v)
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
