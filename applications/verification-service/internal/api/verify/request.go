package verify

import (
	"github.com/go-playground/validator/v10"

	"verification-service/internal/library/server"
)

// Body represents the handler's structured request-body
type Body struct {
	Code string `json:"verification-code" validate:"required"` // Email represents the user's required email address.
}

func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"code": {
			Value:   b.Code,
			Valid:   b.Code != "",
			Message: "(Required) A valid verification code.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies server.Helper
var _ server.Helper = (*Body)(nil)
