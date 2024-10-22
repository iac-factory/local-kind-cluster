package login

import (
	"github.com/go-playground/validator/v10"

	"authentication-service/internal/server"
)

// Body represents the handler's structured request-body
type Body struct {
	Email    string `json:"email" validate:"required,email"`           // Email represents the user's required email address.
	Password string `json:"password" validate:"required,min=8,max=72"` // Password represents the user's required password.
}

func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"password": {
			Valid:   len(b.Password) >= 8 && len(b.Password) <= 72,
			Message: "(Required) The user's password.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies server.Helper
var _ server.Helper = (*Body)(nil)
