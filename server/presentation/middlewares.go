package presentation

import (
	"context"
	"encoding/json"
	"guilhermec-costa/go-web-crawler/server/app"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenSecret := os.Getenv(app.JWTSECRET)
		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}))

		if err != nil || !parsedToken.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "bad jwt claims", http.StatusUnauthorized)
			return
		}

		userId, err := claims.GetSubject()
		if err != nil {
			http.Error(w, "invalid subject", http.StatusUnauthorized)
			return
		}

		ctxWithUserId := context.WithValue(r.Context(), "userId", userId)
		r = r.Clone(ctxWithUserId)
		next.ServeHTTP(w, r)
	})
}

type HandlerFunc func(r *http.Request) (ControllerResponse, int)

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
