package versioning

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"ethr.gg/str"

	"ethr.gg/server/internal/keystore"
	"ethr.gg/server/logging"
)

var implementation = generic{}

type generic struct {
	keystore.Valuer[string]
	
	options Settings
}

func (generic) Value(ctx context.Context) Version {
	if v, ok := ctx.Value(key).(Version); ok {
		return v
	}

	return Version{API: "N/A", Service: "N/A"}
}

func (g generic) Configuration(options ...Variadic) Implementation {
	var settings = g.options

	for _, option := range options {
		option(&settings)
	}

	g.options = settings

	return g
}

func (g generic) Middleware(next http.Handler) http.Handler {
	var name = str.Title(key.String(), func(o str.Options) {
		o.Log = true
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := g.options.Version

			api, service := value.API, value.Service

			slog.Log(ctx, logging.Trace, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.Group(string(key), slog.String("api", api), slog.String("service", service))))

			ctx = context.WithValue(ctx, key, value)

			w.Header().Set("X-API-Version", api)
			w.Header().Set("X-Service-Version", service)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
