package login

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/library/middleware"

	"authentication-service/internal/library/server/cookies"

	"authentication-service/internal/library/server"

	"authentication-service/internal/database"
	"authentication-service/internal/token"
	"authentication-service/models/users"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "login"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

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

	slog.InfoContext(ctx, "Input", slog.Any("body", input))

	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer connection.Release()

	count, e := users.New().Count(ctx, connection, input.Email)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		slog.ErrorContext(ctx, "Unable to Check User Count", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count == 0 {
		slog.WarnContext(ctx, "User Not Found", slog.String("email", input.Email))
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	user, e := users.New().Get(ctx, connection, input.Email)
	if e != nil {
		const message = "Unable to Retrieve User Record"

		slog.WarnContext(ctx, message, slog.String("email", input.Email))
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	if e := users.Verify(user.Password, input.Password); e != nil {
		const message = "Invalid Authentication Attempt"

		slog.WarnContext(ctx, message, slog.String("email", input.Email))
		http.Error(w, message, http.StatusUnauthorized)
		return
	}

	// --> attempt to get customer id

	jwtstring, e := token.Create(ctx, user.Email)
	if e != nil {
		const message = "Unable to Generate JWT Token"

		slog.WarnContext(ctx, message, slog.String("email", input.Email))
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	cookies.Secure(w, "token", jwtstring)

	slog.DebugContext(ctx, "Successfully Generated JWT", slog.String("jwt", jwtstring))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jwtstring))

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
