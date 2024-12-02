package expiration

import (
	"github.com/go-playground/validator/v10"

	"reconnaissance-service/internal/library/server"
)

// Body represents the handler's structured request-body
type Body struct {
	Hostname string `json:"hostname" validate:"required,hostname"` // Hostname represents the target system's hostname, according to RFC 952.
	Port     int    `json:"port" validate:"min=1,max=65535"`       // Port represents the target system's TLS-exposed port, the partial used to construct an address according to RFC 1123.
}

func (b *Body) Help() server.Validators {
	var mapping = server.Validators{
		"hostname": {
			Value:   b.Hostname,
			Valid:   b.Hostname != "",
			Message: "(Required) A valid, RFC-952 specification defined hostname is required.",
		},
		"port": {
			Valid:   b.Port > 0 && b.Port <= 65535,
			Message: "(Required) The system's hostname-related port is required. Port must be in range 0 < port <= 65535.",
		},
	}

	return mapping
}

// v represents the request body's struct validator
var v = validator.New(validator.WithRequiredStructEnabled())

// --> ensure Body satisfies server.Helper
var _ server.Helper = (*Body)(nil)
