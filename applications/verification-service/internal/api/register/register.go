package register

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"verification-service/internal/database"
	"verification-service/internal/library/mail"
	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/authentication"
	"verification-service/internal/library/random"
	"verification-service/models/verifications"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "register"

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

	tx, e := connection.Begin(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer database.Disconnect(ctx, connection, tx)

	// --> check if record exists
	count, e := verifications.New().Count(ctx, tx, email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check Verification Count", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count >= 1 {
		slog.WarnContext(ctx, "Verification Record Already Exists", slog.String("email", email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, "User Already Exists", http.StatusConflict)
		return
	}

	// --> create the new record
	record := &verifications.CreateParams{Email: email, Code: random.Verification()}

	slog.DebugContext(ctx, "Creating New Verification Record", slog.Any("record", record))

	result, e := verifications.New().Create(ctx, tx, record)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Create New Verification Record", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "Sending Email", slog.String("recipient", record.Email))
	if e := mail.Verification(ctx, record.Email, record.Code); e != nil {
		slog.ErrorContext(ctx, "Unable to Send Email", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "Successfully Submitted Email")

	// --> commit the transaction only after all error cases have been evaluated
	if e := tx.Commit(ctx); e != nil {
		slog.ErrorContext(ctx, "Unable to Commit Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Successfully Created Verification Record")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
