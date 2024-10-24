package versioning

import (
	"context"
	"net/http"
)

type Version struct {
	Service string `json:"service" yaml:"service"`
	// API represents an HTTP Server's API version. This value comes from a "X-API-Version" that's often hijacked by Istio in the user's request.
	API string `json:"api" yaml:"api"`
}

type Implementation interface {
	Value(ctx context.Context) Version
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return &generic{
		options: settings(),
	}
}
