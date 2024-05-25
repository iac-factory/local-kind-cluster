package path

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

		{
			value := r.URL.Path

			slog.Log(ctx, logging.Trace, fmt.Sprintf("Evaluating %s Middleware", name), slog.Group("context", slog.String(string(key), value)))

			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
