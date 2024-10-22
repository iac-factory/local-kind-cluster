package registration

import (
	"github.com/go-playground/validator/v10"

	"authentication-service/internal/server"
)

// Body represents the handler's structured request-body/
type Body struct {
	Email    string `json:"email" validate:"required,email"`           // Email represents the user's required email address.
	Password string `json:"password" validate:"required,min=8,max=72"` // Password represents the user's required password.
}

// Help returns a [server.Validators] mapping that intends to be json-encoded to display helpful request-body context requirements.
func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"password": {
			Valid:   len(b.Password) >= 8 && len(b.Password) <= 72,
			Message: "(Required) The user's password. Password must be between 8 and 72 characters in length.",
		},
	}

	return mapping
}

// v represents the request body's struct validator.
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies the server.Helper interface.
var _ server.Helper = (*Body)(nil)
