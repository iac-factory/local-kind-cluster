package logout

import (
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/library/middleware"

	"authentication-service/internal/library/server/cookies"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "logout"

	ctx := r.Context()

	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	cookies.Delete(w, "token")

	redirect := os.Getenv("FRONTEND_URL")
	if redirect == "" {
		slog.WarnContext(ctx, "Front-End Redirect URL for Logout Endpoint Not Specified. Defaulting to \"/\"")

		redirect = "/"
	} else {
		slog.InfoContext(ctx, "Front-End Redirect URL", slog.String("value", redirect))
	}

	slog.DebugContext(ctx, "Logout", slog.String("redirect", redirect))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	http.Redirect(w, r, redirect, http.StatusFound)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
