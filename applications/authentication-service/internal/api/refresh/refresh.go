package refresh

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/authentication"

	"authentication-service/internal/library/server/cookies"

	"authentication-service/internal/token"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "refresh"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	authentication := authentication.New().Value(ctx)

	expiration, e := authentication.Token.Claims.GetExpirationTime()
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		span.RecordError(e, trace.WithStackTrace(true))

		const message = "Unable to Get JWT Expiration Time"

		slog.ErrorContext(ctx, message, slog.String("error", e.Error()))
		http.Error(w, message, http.StatusUnauthorized)
		return
	}

	remaining := time.Until(time.Unix(expiration.Unix(), 0))
	if remaining > 15*time.Minute {
		w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
		w.Header().Set("X-Retry-After-Unit", "Seconds")

		const message = "Refresh Token Requested Too Soon"

		slog.WarnContext(ctx, message, slog.Duration("duration", remaining))
		http.Error(w, message, http.StatusTooManyRequests)
		return
	}

	email := authentication.Email

	update, e := token.Create(ctx, email)
	if e != nil {
		const message = "Error Creating JWT Token"

		slog.WarnContext(ctx, message, slog.String("error", e.Error()))
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	cookies.Secure(w, "token", update)

	slog.DebugContext(ctx, "Successfully Generated JWT Token")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(update))

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
