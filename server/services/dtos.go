package services

type SignInDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (payload SignInDTO) Validate() error {
	return nil
}

type SignUpDTO struct {
	Email string `json:"email"`
	Password string `json:"password"`
}
