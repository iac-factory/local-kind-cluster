package avatar

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/library/levels"

	"user-service/internal/library/middleware"

	"user-service/internal/library/server"

	"user-service/internal/api/avatar/types/update"
	"user-service/internal/database"
	"user-service/models/users"
)

func patch(w http.ResponseWriter, r *http.Request) {
	const name = "avatar-update"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	var input update.Body
	if validator, e := server.Validate(ctx, update.V, r.Body, &input); e != nil {
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

	count, e := users.New().Count(ctx, tx, input.Email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check if User Exists", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count == 0 {
		slog.ErrorContext(ctx, "User Not Found", slog.String("email", input.Email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, "Account with Email Address Not Found", http.StatusNotFound)
		return
	}

	arguments := &users.UpdateUserAvatarParams{Email: input.Email, Avatar: &input.Avatar, Modification: pgtype.Timestamptz{Valid: true, Time: time.Now().UTC()}}
	if e := users.New().UpdateUserAvatar(ctx, tx, arguments); e != nil {
		const message = "Unable to Update User's Avatar"

		slog.ErrorContext(ctx, message, slog.String("email", input.Email), slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusInternalServerError)
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

	slog.Log(ctx, levels.Trace, "Successfully Committed Database Transaction")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(arguments)

	return
}

var Patch = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	patch(w, r)

	return
})
