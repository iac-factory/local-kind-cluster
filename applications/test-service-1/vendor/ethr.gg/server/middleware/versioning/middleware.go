package versioning

import (
	"context"
	"net/http"
)

type Version struct {
	API     string
	Service string
}

type Implementation interface {
	Value(ctx context.Context) Version
	Configuration(options ...Variadic) Implementation
	Middleware(next http.Handler) http.Handler
}

func New() Implementation {
	return implementation
}
