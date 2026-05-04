package services

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (payload LoginDTO) Validate() error {
	return nil
}
