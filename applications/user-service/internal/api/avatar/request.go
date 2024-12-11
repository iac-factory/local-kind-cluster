package avatar

import (
	"github.com/go-playground/validator/v10"

	"user-service/internal/library/server"
)

// Body represents the handler's structured request-body
type Body struct {
	server.Helper `json:"-"`

	Avatar string `json:"avatar" validate:"required,http_url"`
}

func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"avatar": {
			Value:   b.Avatar,
			Valid:   b.Avatar != "",
			Message: "(Required) An avatar URL.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies server.Helper
var _ server.Helper = (*Body)(nil)
