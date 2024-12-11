package registration

import (
	"github.com/go-playground/validator/v10"

	"user-service/internal/library/server"
)

// Body represents the handler's structured request-body
type Body struct {
	server.Helper `json:"-"`

	Email  string  `json:"email" validate:"required,email"` // Email represents the user's required email address.
	Avatar *string `json:"avatar"`                          // Avatar represents an optional field for a valid HTTP URL pointing to the user's avatar.
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
			Valid:   true,
			Message: "(Optional) An avatar URL.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies server.Helper
var _ server.Helper = (*Body)(nil)
