package telemetry

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Instance struct {
	Client *http.Client

	Headers map[string]string
}

func Client(headers map[string]string) *Instance {
	return &Instance{
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
		Headers: headers,
	}
}

func (c *Instance) Do(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("polygun-http-telemetry-client").Start(ctx, r.URL.String())

	defer span.End()

	slog.DebugContext(ctx, "Log Message From Polygun HTTP Client Transport", slog.String("url", r.URL.String()))

	for key, value := range c.Headers {
		r.Header.Set(key, value)
	}

	return c.Client.Do(r)
}
