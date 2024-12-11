package tracing

import (
	"go.opentelemetry.io/otel/trace"

	"verification-service/internal/library/middleware/types"
)

type Settings struct {
	Tracer trace.Tracer
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}