package services

import (
	"crypto/sha256"
	"fmt"
	"guilhermec-costa/go-web-crawler/server/infra"
)

func SignInHandler(payload SignInDTO, store infra.UserDAO) error {
	return nil
}

func SignUpHandler(payload SignUpDTO, store infra.UserDAO) (int64, error) {
	h := sha256.New()
	h.Write([]byte(payload.Password))
	pwdHash := h.Sum(nil)

	result, err := store.Create(payload.Email, string(pwdHash))
	if err != nil {
		return 0, fmt.Errorf("Failed to create user: %v", err)
	}

	return result, nil
}
