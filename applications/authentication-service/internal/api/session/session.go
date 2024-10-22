package session

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/middleware"

	"authentication-service/internal/middleware/authentication"

	"authentication-service/internal/database"
	"authentication-service/models/users"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "session"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	// --> retrieve authentication context

	authentication := authentication.New().Value(ctx)

	email := authentication.Email

	// --> establish database connection

	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer connection.Release()

	// --> extract the full user record

	record, e := users.New().Get(ctx, connection, email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Extract User Database Record", slog.Any("error", e))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, "Unable to Extract User Database Record", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Successfully Extracted User Record for Session", slog.Any("user", record))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
