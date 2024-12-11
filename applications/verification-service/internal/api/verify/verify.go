package verify

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"verification-service/internal/library/middleware/authentication"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/server"

	"verification-service/internal/database"
	"verification-service/models/verifications"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "verify"

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

	// --> verify input
	var input Body
	if validator, e := server.Validate(ctx, v, r.Body, &input); e != nil {
		slog.WarnContext(ctx, "Unable to Verify Request Body")

		if validator != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(validator)

			return
		}

		http.Error(w, "Unable to Validate Request Body", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Input", slog.Any("body", input))

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
	} else if count == 0 {
		slog.WarnContext(ctx, "Verification Record Not Found", slog.String("email", email))
		http.Error(w, "Verification Record Not Found", http.StatusNotFound)
		return
	}

	// --> validate the token(s)
	verification, e := verifications.New().Get(ctx, connection, email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Get Verification Record", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	if verification.Modification.Valid && now.After(verification.Modification.Time.Add(time.Hour*24)) {
		slog.ErrorContext(ctx, "Expired Verification Request")
		http.Error(w, "Expired Verification Request Token", http.StatusGone)
		return
	} else if now.After(verification.Creation.Time.Add(time.Hour * 24)) {
		slog.ErrorContext(ctx, "Expired Verification Request")
		http.Error(w, "Expired Verification Request Token", http.StatusGone)
		return
	} else if verification.Verified {
		slog.ErrorContext(ctx, "User Already Verified")
		http.Error(w, "User Already Verified", http.StatusUnprocessableEntity)
		return
	}

	if input.Code != verification.Code {
		slog.ErrorContext(ctx, "Invalid Verification Request")
		http.Error(w, "Invalid Verification Request Token", http.StatusConflict)
		return
	}

	slog.DebugContext(ctx, "Verifying Verification Record")

	if e := verifications.New().Verify(ctx, connection, &verifications.VerifyParams{Email: email, Modification: pgtype.Timestamptz{Valid: true, Time: time.Now().UTC()}}); e != nil {
		slog.ErrorContext(ctx, "Unable to Verify Verification Record", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Successfully Verified User")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
