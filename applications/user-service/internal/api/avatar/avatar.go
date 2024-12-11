package avatar

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/database"
	"user-service/internal/library/middleware"
	"user-service/internal/library/middleware/authentication"
	"user-service/internal/library/server"
	"user-service/models/users"
)

// Handler is an HTTP handler that processes avatar updates for authenticated users, ensuring authorization and database consistency.
var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	const name = "avatar"

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

	slog.DebugContext(ctx, "Executing Avatar Handler", slog.String("email", email))

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

	slog.DebugContext(ctx, "Input", slog.Any("request", input))

	// Get the user's identifier.
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Typecast ID from Path Value to Int64", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Establish connection to database.
	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Initialize a database transaction in the event of rollback.
	tx, e := connection.Begin(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer database.Disconnect(ctx, connection, tx)

	// Check if the database record exists.
	exists, e := users.New().Exists(ctx, tx, id)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check if User Exists", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if !(exists) {
		slog.WarnContext(ctx, "Active User Record Not Found", slog.String("email", email), slog.Int64("id", id))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User Record Not Found",
		})

		return
	}

	// Ensure database record's email-address matches authenticated user.

	{
		row, e := users.New().GetUserEmailAddressByID(ctx, tx, id)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Query User for Email & ID Information",
				slog.Int64("id", id),
				slog.String("email", email),
				slog.String("error", e.Error()),
				slog.String("error-type", reflect.TypeOf(e).String()),
			)

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if email != row.Email {
			slog.ErrorContext(ctx, "Potential Hijack, Token Forgery Event",
				slog.Int64("id", id),
				slog.String("authentication-email", email),
				slog.String("database-record-email", row.Email),
			)

			labeler.Add(attribute.Bool("error", true))
			labeler.Add(attribute.Bool("security-risk", true))

			exception := server.Exception{
				Code:    http.StatusForbidden,
				Message: "Email mismatch: You are not authorized to update the user-avatar.",
				Internal: &server.Internal{
					Error:   fmt.Errorf("email mismatch: user is not authorized to update the target database record"),
					Message: "Potential Database Inconsistency, Hijack, or Token Forgery Event Detected",
				},
			}

			exception.Response(w)
			return
		}
	}

	if e := users.New().UpdateUserAvatar(ctx, tx, &users.UpdateUserAvatarParams{Avatar: input.Avatar, ID: id}); e != nil {
		slog.ErrorContext(ctx, "Unable to Update User's Avatar",
			slog.Int64("id", id),
			slog.String("email", email),
			slog.String("error", e.Error()),
			slog.String("error-type", reflect.TypeOf(e).String()),
		)

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// --> commit the transaction
	if e := tx.Commit(ctx); e != nil {
		const message = "Unable to Commit Transaction"

		slog.ErrorContext(ctx, message, slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "Successfully Updated User's Avatar", slog.String("email", email), slog.Int64("id", id))

	w.WriteHeader(http.StatusNoContent)
	return
})
