package remove

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"authentication-service/internal/middleware"

	"authentication-service/internal/middleware/authentication"

	"authentication-service/internal/database"
	"authentication-service/models/users"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "remove"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	// --> retrieve authentication context

	authentication := authentication.New().Value(ctx)

	email := authentication.Email

	// --> get record-id value from path

	var id int64
	i, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Typecast ID from Path Value to Int64", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id = i

	// --> establish database connection and transaction

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

	// --> remove user record

	if e := users.New().Remove(ctx, tx, &users.RemoveParams{Email: email, ID: id}); e != nil {
		slog.ErrorContext(ctx, "Unable to Remove User", slog.Any("error", e))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, "Unable to Remove User", http.StatusInternalServerError)
		return
	}

	// --> commit the transaction only after all error cases have been evaluated

	if e := tx.Commit(ctx); e != nil {
		const message = "Unable to Commit Transaction"

		slog.ErrorContext(ctx, message, slog.String("email", email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusConflict)
		return
	}

	slog.InfoContext(ctx, "Successfully Removed User Record", slog.String("email", email))

	// --> extract the full user record, ensuring usage of only the connection and not the transaction

	record, e := users.New().Extract(ctx, connection, &users.ExtractParams{ID: id, Email: email})
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Extract User Database Record", slog.Any("error", e))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, "Unable to Extract User Database Record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
