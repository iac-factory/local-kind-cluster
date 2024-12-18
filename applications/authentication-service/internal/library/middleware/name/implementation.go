package name

import (
	"context"
	"log/slog"
	"net/http"

	"authentication-service/internal/library/middleware/types"
)

type generic struct {
	types.Valuer[string]

	options *Settings
}

func (g *generic) Configuration(options ...Variadic) Implementation {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	g.options = o

	return g
}

func (*generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (g *generic) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := g.options.Service

			slog.Log(ctx, (slog.LevelDebug - 4), "Middleware", slog.Group("context", slog.String("key", string(key)), slog.String("value", value)))

			ctx = context.WithValue(ctx, key, value)

			w.Header().Set("X-Service-Name", value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
