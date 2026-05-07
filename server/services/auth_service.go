package services

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/server/infra"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	store infra.UserDAO
}

func (a *AuthService) SignInHandler(payload SignInDTO) (string, error) {
	user, err := a.store.FindByEmail(payload.Email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	signedToken, err := token.SignedString([]byte("churros"))
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt token")
	}

	return signedToken, nil
}

func (a *AuthService) SignUpHandler(payload SignUpDTO) (int64, error) {
	slog.Info("creating user", "email", payload.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := a.store.Create(payload.Email, string(hash))
	if err != nil {
		slog.Error("failed to create user", "email", payload.Email)
		return 0, fmt.Errorf("Failed to create user: %v", err)
	}

	slog.Info("user successfully created", "email", payload.Email)
	return result, nil
}

func NewAuthService(s infra.UserDAO) *AuthService {
	return &AuthService{
		store: s,
	}
}
