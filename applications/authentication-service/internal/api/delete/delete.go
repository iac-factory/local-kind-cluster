package delete

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/database"
	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/authentication"
	"authentication-service/internal/library/server"
	"authentication-service/internal/library/server/cookies"
	"authentication-service/models/users"
)

// Handler is an HTTP handler for processing user deletion requests, supporting both soft and hard delete operations with
// authentication and transaction handling.
var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	const name = "delete"

	ctx := r.Context()

	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)
	labeler, _ := otelhttp.LabelerFromContext(ctx)

	defer span.End()

	// Retrieve authentication context.
	token := authentication.New().Value(ctx).Token

	claims := token.Claims.(jwt.MapClaims)

	email, e := claims.GetSubject()
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Get JWT Subject", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

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

	// Determine the type of delete operation is being requested -- defaults to soft.
	operation := strings.ToLower(r.URL.Query().Get("type"))
	if operation == "" {
		operation = "soft"
	}

	slog.DebugContext(ctx, "Delete User Operation", slog.String("type", operation))

	if operation == "hard" {
		// Check if the database record exists (hard).
		exists, e := users.New().ExistsForce(ctx, tx, id)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Check if User Exists", slog.String("error", e.Error()), slog.String("operation", "hard"))
			labeler.Add(attribute.Bool("error", true))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !(exists) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// Ensure database record's email-address matches authenticated user.

		{
			row, e := users.New().GetUserEmailAddressByIDForce(ctx, tx, id)
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
					Message: "Email mismatch: You are not authorized to delete the user record.",
					Internal: &server.Internal{
						Error:   fmt.Errorf("email mismatch: user is not authorized to delete the target database record"),
						Message: "Potential Database Inconsistency, Hijack, or Token Forgery Event Detected",
					},
				}

				exception.Response(w)
				return
			}
		}

		if e := users.New().DeleteHard(ctx, tx, id); e != nil {
			slog.ErrorContext(ctx, "Unable to Delete (Hard) User Record",
				slog.Int64("id", id),
				slog.String("email", email),
				slog.String("error", e.Error()),
				slog.String("error-type", reflect.TypeOf(e).String()),
			)

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, "Unable to Remove User", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Successfully Removed User Record", slog.String("email", email), slog.Int64("id", id), slog.String("operation", "hard"))
	} else { // --> default condition
		// Check if the database record exists (soft).
		exists, e := users.New().Exists(ctx, tx, id)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Check if User Exists", slog.String("error", e.Error()), slog.String("operation", "soft"))
			labeler.Add(attribute.Bool("error", true))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !(exists) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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
					Message: "Email mismatch: You are not authorized to delete the user record.",
					Internal: &server.Internal{
						Error:   fmt.Errorf("email mismatch: user is not authorized to delete the target database record"),
						Message: "Potential Database Inconsistency, Hijack, or Token Forgery Event Detected",
					},
				}

				exception.Response(w)
				return
			}
		}

		if e := users.New().DeleteSoft(ctx, tx, id); e != nil {
			slog.ErrorContext(ctx, "Unable to Delete (Soft) User Record",
				slog.Int64("id", id),
				slog.String("email", email),
				slog.String("error", e.Error()),
				slog.String("error-type", reflect.TypeOf(e).String()),
			)

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, "Unable to Remove User", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Successfully Removed User Record", slog.String("email", email), slog.Int64("id", id), slog.String("operation", "soft"))
	}

	// Commit the transaction only after all error cases have been evaluated.
	if e := tx.Commit(ctx); e != nil {
		const message = "Unable to Commit Transaction"

		slog.ErrorContext(ctx, message, slog.String("email", email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusConflict)
		return
	}

	slog.DebugContext(ctx, "Successfully Removed User Record", slog.String("email", email), slog.Int64("id", id))

	cookies.Delete(w, "token")
	w.WriteHeader(http.StatusNoContent)
	return
})
