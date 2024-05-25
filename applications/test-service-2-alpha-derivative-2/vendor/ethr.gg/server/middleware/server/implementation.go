package server

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
	var name = str.Title(key.String(), func(o str.Options) {
		o.Log = true
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		{
			value := g.options.Server

			slog.Log(ctx, logging.Trace, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(string(key), value)))

			ctx = context.WithValue(ctx, key, value)

			w.Header().Set("Server", value)
			w.Header().Set("X-Server-Identifier", value) // envoy proxy removes server header so the x header is set
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
