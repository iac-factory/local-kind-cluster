package registration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/telemetrics"
	"authentication-service/internal/library/server"
	"authentication-service/internal/library/server/cookies"
	"authentication-service/internal/library/server/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/database"
	"authentication-service/internal/token"
	"authentication-service/models/users"
)

type exception struct {
	code    int
	status  string
	message string
}

func (e *exception) Error() string {
	return fmt.Sprintf("%s: %s", e.status, e.message)
}

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "registration"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	// --> check if authenticated session already is established

	cookie, e := r.Cookie("token")
	if e == nil {
		jwttoken, e := token.Verify(ctx, cookie.Value)
		if e == nil && jwttoken.Valid {
			slog.WarnContext(ctx, "User is Already Authenticated", slog.String("email", jwttoken.Claims.(jwt.MapClaims)["sub"].(string)))

			labeler.Add(attribute.Bool("error", true))
			http.Error(w, "Authenticated Session Already Exists for User", http.StatusBadRequest)
			return
		}
	}

	// --> verify input

	var input Body
	if validator, e := server.Validate(ctx, v, r.Body, &input); e != nil {
		slog.WarnContext(ctx, "Request Body Validation Failure")

		if validator != nil {
			slog.ErrorContext(ctx, "Unable to Verify Request Body (Validator)", slog.Any("validator", validator))

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

	tx, e := connection.Begin(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer database.Disconnect(ctx, connection, tx)

	user := &users.CreateParams{Email: input.Email}

	// --> check if user exists
	count, e := users.New().Count(ctx, tx, user.Email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check User Count", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count >= 1 {
		slog.WarnContext(ctx, "User Already Exists", slog.String("email", input.Email))

		http.Error(w, "User Already Exists", http.StatusConflict)
		return
	}

	password := input.Password
	user.Password, e = users.Hash(password)
	if e != nil {
		slog.ErrorContext(ctx, "Unknown Exception - Unable to Hash User's Password", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, e := users.New().Create(ctx, tx, user)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Create New User", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jwtstring, e := token.Create(ctx, result.Email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Create JWT String", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// register the user with user-service
	var events = func() error { // --> only internal server errors relative to the current service will return an error
		headers := telemetrics.New().Value(ctx).Headers
		maps.Copy(headers, map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", jwtstring),
		})

		c := telemetry.Client(headers)

		var reader bytes.Buffer
		if e := json.NewEncoder(&reader).Encode(map[string]string{"email": user.Email}); e != nil {
			e = fmt.Errorf("unable to encode email address: %w", e)

			slog.ErrorContext(ctx, "Unable to Encode Email", slog.String("error", e.Error()))

			return e
		}

		namespace := os.Getenv("NAMESPACE")
		if namespace == "" {
			namespace = "development"
		}

		url := fmt.Sprintf("http://user-service.%s.svc.cluster.local:8080/register", namespace)
		if override, ok := ctx.Value("user-service-registration-endpoint").(string); ok {
			url = override // currently used for overriding the user-service endpoint during unit-testing
		}

		request, e := http.NewRequestWithContext(ctx, http.MethodPost, url, &reader)
		if e != nil {
			slog.WarnContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))

			return nil
		}

		response, e := c.Do(request)
		if e != nil {
			switch {
			case strings.Contains(e.Error(), "no such host"):
				slog.WarnContext(ctx, "User-Service Registration Endpoint Not Found", slog.String("error", e.Error()))
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

		// rollback conditions
		switch response.StatusCode {
		case http.StatusInternalServerError:
			slog.WarnContext(ctx, "User-Service Registration Endpoint Fatal Error", slog.String("content", string(content)), slog.String("status", response.Status), slog.Int("status-code", response.StatusCode))

			return &exception{code: response.StatusCode, status: response.Status, message: "Internal Server Error"}
		}

		slog.InfoContext(ctx, "User-Service Registration Response", slog.String("content", string(content)), slog.String("status", response.Status), slog.Int("status-code", response.StatusCode))

		return nil
	}

	if e := events(); e != nil {
		labeler.Add(attribute.Bool("error", true))

		switch e.(type) {
		case *exception:
			instance := e.(*exception)

			http.Error(w, instance.status, instance.code)

		default:
			http.Error(w, "Unhandled Exception", http.StatusInternalServerError)
		}

		return
	}

	// --> commit the transaction only after all error cases have been evaluated
	if e := tx.Commit(ctx); e != nil {
		slog.ErrorContext(ctx, "Unable to Commit Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Successfully Created User", slog.Any("user", result))

	cookies.Secure(w, "token", jwtstring)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(jwtstring))

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
