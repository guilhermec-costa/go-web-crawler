package presentation

import (
	"encoding/json"
	"log/slog"
	"net/http"
	// "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("validating jwt token")
		next.ServeHTTP(w, r)
	})
}

type HandlerFunc func(r *http.Request) (any, int)

func JsonHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status := h(r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to encode response", "err", err)
		}
	}
}
