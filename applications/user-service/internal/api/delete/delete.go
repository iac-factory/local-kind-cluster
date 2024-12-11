package delete

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/library/middleware"
	"user-service/internal/library/middleware/authentication"
	"user-service/internal/library/middleware/telemetrics"
	"user-service/internal/library/server"
	"user-service/internal/library/server/cookies"
	"user-service/internal/library/server/telemetry"

	"user-service/internal/database"
	"user-service/models/users"
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

	// Remove the user in authentication-service
	var events = func() error { // --> only internal server errors relative to the current service will return an error
		headers := telemetrics.New().Value(ctx).Headers
		maps.Copy(headers, map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token.Raw),
		})

		c := telemetry.Client(headers)

		var reader bytes.Buffer
		if e := json.NewEncoder(&reader).Encode(map[string]string{"email": email}); e != nil {
			e = fmt.Errorf("unable to encode email address: %w", e)

			slog.ErrorContext(ctx, "Unable to Encode Email", slog.String("error", e.Error()))

			return e
		}

		// Get the authentication-service user's identifier.
		var identifier float64 // --> JSON integers always unmarshal to float64

		{
			force := "false"
			if operation == "hard" {
				force = "true"
			}

			url := fmt.Sprintf("%s://%s:%d/session?force=%s", "http", "authentication-service", 8080, force)
			if override, ok := ctx.Value("authentication-service-session-endpoint").(string); ok {
				url = override // currently used for overriding the user-service endpoint during unit-testing
			}

			slog.DebugContext(ctx, "Authentication-Service Session URL", slog.String("url", url))

			request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if e != nil {
				slog.WarnContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))

				return nil
			}

			response, e := c.Do(request)
			if e != nil {
				switch {
				case strings.Contains(e.Error(), "no such host"):
					slog.WarnContext(ctx, "Authentication-Service User-Session Endpoint Not Found", slog.String("error", e.Error()))
					// --> occurs during local testing due to lack of internal kubernetes networking
					return nil
				default:
					slog.WarnContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))

					return e
				}
			}

			defer response.Body.Close()

			content, e := io.ReadAll(response.Body)
			if e != nil {
				slog.WarnContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))

				return nil
			}

			if response.StatusCode != http.StatusOK {
				switch response.StatusCode {
				default:
					slog.WarnContext(ctx, "Authentication-Service Session Endpoint Fatal Error - Unhandled Status-Code", slog.String("content", string(content)), slog.String("status", response.Status), slog.Int("status-code", response.StatusCode))

					exception := server.Exception{Code: response.StatusCode, Status: response.Status}

					return &exception
				}
			}

			slog.InfoContext(ctx, "User-Service Registration Response", slog.String("content", string(content)), slog.String("status", response.Status), slog.Int("status-code", response.StatusCode))

			var datum map[string]interface{}
			if e := json.Unmarshal(content, &datum); e != nil {
				slog.ErrorContext(ctx, "Unable to Unmarshal JSON", slog.String("error", e.Error()))

				return e
			}

			v, ok := datum["id"]
			if !ok {
				slog.ErrorContext(ctx, "Unable to Get ID Information from Authentication-Service Session Endpoint", slog.String("key", "id"), slog.Any("map", datum))

				return fmt.Errorf("unable to get id information from authetnication-service session endpoint")
			}

			typecast, valid := v.(float64)
			if !valid {
				slog.ErrorContext(ctx, "Unable to Typecast ID to Integer from Authentication-Service Session Endpoint", slog.Any("value", v), slog.Any("response", datum), slog.String("type", reflect.TypeOf(v).String()))

				return fmt.Errorf("unable to get id information from authetnication-service session endpoint")
			}

			identifier = typecast
		}

		url := fmt.Sprintf("%s://%s:%d/users/%d?type=%s", "http", "authentication-service", 8080, int64(identifier), operation)
		if override, ok := ctx.Value("authentication-service-user-delete-endpoint").(string); ok {
			url = override // currently used for overriding the user-service endpoint during unit-testing
		}

		slog.DebugContext(ctx, "Authentication-Service Delete-User URL", slog.String("url", url))

		request, e := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
		if e != nil {
			slog.WarnContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))

			return nil
		}

		response, e := c.Do(request)
		if e != nil {
			switch {
			case strings.Contains(e.Error(), "no such host"):
				slog.WarnContext(ctx, "Authentication-Service User-Delete Endpoint Not Found", slog.String("error", e.Error()))
				// --> occurs during local testing due to lack of internal kubernetes networking
				return nil
			default:
				slog.WarnContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))

				return e
			}
		}

		defer response.Body.Close()

		content, e := io.ReadAll(response.Body)
		if e != nil {
			slog.WarnContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))

			return nil
		}

		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
			switch response.StatusCode {
			default:
				slog.WarnContext(ctx, "Authentication-Service User-Delete Endpoint Fatal Error - Unhandled Status-Code", slog.String("content", string(content)), slog.String("status", response.Status), slog.Int("status-code", response.StatusCode))

				exception := server.Exception{Code: response.StatusCode, Status: response.Status}

				return &exception
			}
		}

		return nil
	}

	if e := events(); e != nil {
		labeler.Add(attribute.Bool("error", true))

		var exception *server.Exception
		if errors.As(e, &exception) {
			slog.ErrorContext(ctx, "Authentication-Service Event Error", slog.String("error", e.Error()))
			exception.Response(w)
			return
		}

		slog.ErrorContext(ctx, "Authentication-Service Unhandled Event Error", slog.String("error", e.Error()))

		http.Error(w, "Unhandled Exception", http.StatusInternalServerError)

		return
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
