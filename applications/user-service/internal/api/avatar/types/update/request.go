package update

import (
	"github.com/go-playground/validator/v10"

	"user-service/internal/library/server"
)

// Body represents the handler's structured request-body
type Body struct {
	server.Helper `json:"-"`
	Email         string `json:"email" validate:"required,email"` // Email represents the user's required email address.
	Avatar        string `json:"avatar" validate:"required"`
}

func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"email": {
			Value:   b.Email,
			Valid:   b.Email != "",
			Message: "(Required) A valid, unique email address.",
		},
		"avatar": {
			Value:   b.Avatar,
			Valid:   b.Avatar != "",
			Message: "(Required) An avatar URL.",
		},
	}

	return mapping
}

// V represents the request body's struct validator
var V = validator.New(validator.WithRequiredStructEnabled())
