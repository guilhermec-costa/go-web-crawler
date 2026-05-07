package presentation

import (
	"context"
	"encoding/json"
	"guilhermec-costa/go-web-crawler/server/app"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type UserContextKey string

func unauthorizedResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(ControllerResponse{
		Error: "missing authorization header",
	})
}

const userContextKey UserContextKey = "userId"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		SetJSONContentType(w)

		if authHeader == "" {
			unauthorizedResponse(w, "missing authorization header")
			return
		}

		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			unauthorizedResponse(w, "invalid authorization format")
			return
		}

		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			return []byte(app.JWTSECRET), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || !parsedToken.Valid {
			unauthorizedResponse(w, "invalid token")
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			unauthorizedResponse(w, "bad jwt claims")
			return
		}

		userId, err := claims.GetSubject()
		if err != nil {
			unauthorizedResponse(w, "invalid subject")
			return
		}

		ctxWithUserId := context.WithValue(r.Context(), userContextKey, userId)
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
