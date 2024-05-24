package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
	const name = "Telemetry"

	var path = middleware.Keys().Path()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		pattern := ctx.Value(middleware.Keys().Path()).(string)

		{
			value := "enabled"

			slog.DebugContext(ctx, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(path.String(), pattern), slog.String(key.String(), value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next = otelhttp.WithRouteTag(pattern, next)

		next.ServeHTTP(w, r)
	})
}
