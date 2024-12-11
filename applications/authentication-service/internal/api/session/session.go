package session

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/authentication"

	"authentication-service/internal/database"
	"authentication-service/models/users"
)

// Handler processes incoming HTTP requests, retrieves user session data, establishes a database connection, fetches the user record, and returns it in JSON format.
var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	const name = "session"

	ctx := r.Context()

	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)
	labeler, _ := otelhttp.LabelerFromContext(ctx)

	defer span.End()

	// Retrieve authentication context.
	claims := authentication.New().Value(ctx).Token.Claims.(jwt.MapClaims)

	email, e := claims.GetSubject()
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Get JWT Subject", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Establish database connection.
	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer connection.Release()

	// Check if a hard search should be performed.
	force := strings.ToLower(r.URL.Query().Get("force"))
	if force == "" {
		force = "false"
	}

	if force == "true" {
		// Extract the full user record.
		record, e := users.New().GetForce(ctx, connection, email)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Extract User Database Record", slog.String("force", force), slog.Any("error", e), slog.String("error-type", reflect.TypeOf(e).String()))

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, "Unable to Extract User Database Record", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Successfully Extracted User Record for Session", slog.Any("user", record), slog.String("force", force))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(record)
	} else {
		// Extract the full user record.
		record, e := users.New().Get(ctx, connection, email)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Extract User Database Record", slog.String("force", force), slog.Any("error", e), slog.String("error-type", reflect.TypeOf(e).String()))

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, "Unable to Extract User Database Record", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Successfully Extracted User Record for Session", slog.Any("user", record), slog.String("force", force))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(record)
	}

	return
})
