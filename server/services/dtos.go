package services

import (
	"fmt"
	"net/mail"
)

type SignInDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p SignInDTO) Validate() error {
	_, err := mail.ParseAddress(p.Email)
	if err != nil {
		return fmt.Errorf("Invalid email format")
	}

	return nil
}

type SignUpDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p SignUpDTO) Validate() error {
	_, err := mail.ParseAddress(p.Email)
	if err != nil {
		return fmt.Errorf("Invalid email format")
	}

	if len(p.Password) < 8 {
		return fmt.Errorf("Password must have at least 8 characters")
	}

	return nil
}
