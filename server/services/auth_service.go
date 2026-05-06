package services

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/server/infra"
	"log/slog"
	"golang.org/x/crypto/bcrypt"
)

func SignInHandler(payload SignInDTO, store infra.UserDAO) error {
	return nil
}

func SignUpHandler(payload SignUpDTO, store infra.UserDAO) (int64, error) {
	slog.Info("creating user", "email", payload.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := store.Create(payload.Email, string(hash))
	if err != nil {
		slog.Error("failed to create user", "email", payload.Email)
		return 0, fmt.Errorf("Failed to create user: %v", err)
	}

	slog.Info("user successfully created", "email", payload.Email)
	return result, nil
}
