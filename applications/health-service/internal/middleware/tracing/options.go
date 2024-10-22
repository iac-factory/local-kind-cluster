package tracing

import (
	"go.opentelemetry.io/otel/trace"

	"health-service/internal/middleware/types"
)

type Settings struct {
	Tracer trace.Tracer
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
