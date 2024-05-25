package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"ethr.gg/str"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"ethr.gg/server/internal/keystore"
	"ethr.gg/server/logging"
	"ethr.gg/server/middleware/path"
)

var implementation = generic{}

type generic struct {
	keystore.Valuer[string]
}

func (generic) Value(ctx context.Context) string {
	return ctx.Value(key).(string)
}

func (generic) Middleware(next http.Handler) http.Handler {
	var name = str.Title(key.String(), func(o str.Options) {
		o.Log = true
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		route := path.New().Value(ctx)

		pattern := ctx.Value(keystore.Keys().Path()).(string)

		{
			value := "enabled"

			slog.Log(ctx, logging.Trace, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(route, pattern), slog.String(key.String(), value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next = otelhttp.WithRouteTag(pattern, next)

		next.ServeHTTP(w, r)
	})
}
