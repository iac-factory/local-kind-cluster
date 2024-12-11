package status

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/authentication"

	"verification-service/internal/database"
	"verification-service/internal/library/mail"
	"verification-service/internal/library/random"
	"verification-service/models/verifications"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "status"

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

	// --> construct database payload & establish connection, transaction
	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer connection.Release()

	// --> check if record exists
	count, e := verifications.New().Count(ctx, connection, email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check Verification Count", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count == 0 { // --> in efforts in improving data resilience, it can be assumed that if a user is
		// successfully authenticated, but does not contain a database record, then one can
		// simply be created

		// --> create the new record
		slog.DebugContext(ctx, "Creating New Verification Record")

		code := random.Verification()

		if e := mail.Verification(ctx, email, code); e != nil {
			slog.ErrorContext(ctx, "Unable to Send Email", slog.String("error", e.Error()))

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		slog.DebugContext(ctx, "Successfully Submitted Email")

		result, e := verifications.New().Create(ctx, connection, &verifications.CreateParams{Email: email, Code: code})
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Create New Verification Record", slog.String("error", e.Error()))

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		slog.DebugContext(ctx, "Successfully Established Verification Database Record", slog.Any("record", result))
	}

	// --> validate the token(s)
	verification, e := verifications.New().Status(ctx, connection, email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Get Verification Record", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := map[string]bool{"verified": verification.Verified}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
