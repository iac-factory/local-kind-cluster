package path

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"ethr.gg/server/internal/middleware"
)

var implementation = generic{}

type generic struct {
	middleware.Valuer[string]
}

func (generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (generic) Middleware(next http.Handler) http.Handler {
	const name = "Path"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := r.URL.Path

			slog.DebugContext(ctx, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(key.String(), value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
