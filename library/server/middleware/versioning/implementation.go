package versioning

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"ethr.gg/server/internal/middleware"
)

var implementation = generic{}

type generic struct {
	options Settings

	middleware.Valuer[string]
}

func (g generic) Configuration(options ...Variadic) Implementation {
	var settings = g.options

	for _, option := range options {
		option(&settings)
	}

	g.options = settings

	return g
}

func (generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (g generic) Middleware(next http.Handler) http.Handler {
	const name = "Version"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := g.options.Version

			slog.DebugContext(ctx, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(key.String(), value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
