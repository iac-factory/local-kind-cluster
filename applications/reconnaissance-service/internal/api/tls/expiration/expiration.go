package expiration

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"reconnaissance-service/internal/library/middleware"

	"reconnaissance-service/internal/library/server"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "expiration"

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

	// Construct the address variable that's to be used for generating the tls connection
	address := fmt.Sprintf("%s:%d", input.Hostname, input.Port)

	// Check if hostname exists

	// Attempt to resolve the hostname to an IP address
	_, e := net.LookupIP(input.Hostname)
	if e != nil {
		labeler.Add(attribute.Bool("warning", true))
		slog.WarnContext(ctx, "Hostname Doesn't Exist or Cannot be Resolved", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Establish a TLS connection to the hostname
	connection, e := tls.Dial("tcp", address, nil)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		slog.ErrorContext(ctx, "Unable to Establish TLS Connection", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Defer closing of the TLS connection, but with a wrapper for unexpected error-handling purposes
	defer func(connection *tls.Conn) {
		e := connection.Close()
		if e != nil {
			labeler.Add(attribute.Bool("warning", true))
			slog.WarnContext(ctx, "Unable to Close TLS Connection", slog.String("error", e.Error()))
		}
	}(connection)

	// Retrieve the TLS connection state
	state := connection.ConnectionState()

	// Check if any certificates are present
	if len(state.PeerCertificates) == 0 {
		message := fmt.Sprintf("No Peer Certificates Found for Address %s", address)

		labeler.Add(attribute.Bool("warning", true))
		slog.WarnContext(ctx, "No Peer Certificates Found", slog.String("address", address), slog.String("message", message))
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}

	// Get the leaf certificate (the server's certificate)
	cert := state.PeerCertificates[0]

	// Extract the expiration date
	expiration := cert.NotAfter

	// Calculate remaining validity period
	remaining := time.Until(expiration)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"expiration": map[string]interface{}{
			"string": expiration.String(),
			"utc":    expiration.UTC().String(),
			"unix":   expiration.Unix(),
		},
		"time-remaining": map[string]interface{}{
			"string":       remaining.String(),
			"milliseconds": remaining.Milliseconds(),
			"seconds":      remaining.Seconds(),
			"nanoseconds":  remaining.Nanoseconds(),
			"hours":        remaining.Hours(),
		},
	})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
