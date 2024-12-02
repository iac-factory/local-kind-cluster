package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"reconnaissance-service/internal/library/middleware"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Expiration extracts expiration information from a given hostname and port.
func Expiration(ctx context.Context, hostname string, port int) {
	const name = "datum-tls-expiration"

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	address := fmt.Sprintf("%s:%d", hostname, port)

	// Establish a TLS connection to the hostname
	connection, e := tls.Dial("tcp", address, nil)
	if e != nil {
		log.Fatalf("Failed to connect to %s:%s: %v", hostname, port, e)
	}

	var closure = func(callable func() error) {
		if e := callable(); e != nil {
			slog.ErrorContext(ctx, "Received an Error During Closure", slog.String("error", e.Error()))
			labeler.Add(attribute.Bool("error", true))
			return
		}
	}

	defer func(connection *tls.Conn) {
		err := connection.Close()
		if err != nil {

		}
	}(connection)

	// Retrieve the TLS connection state
	state := connection.ConnectionState()

	// Check if any certificates are present
	if len(state.PeerCertificates) == 0 {
		log.Fatalf("No certificates found for %s", hostname)
	}

	// Get the leaf certificate (the server's certificate)
	cert := state.PeerCertificates[0]

	// Extract the expiration date
	expiration := cert.NotAfter

	// Calculate remaining validity period
	remaining := time.Until(expiration)

	// Print the expiration date and remaining days
	fmt.Printf("The TLS/SSL certificate for %s expires on: %s\n", hostname, expiration.Format(time.RFC1123))
	fmt.Printf("The certificate is valid for another %d days\n", int(remaining.Hours()/24))

	// Alert if the certificate is expired or near expiry
	if remaining <= 0 {
		fmt.Println("Alert: The certificate has expired!")
	} else if remaining < (30 * 24 * time.Hour) {
		fmt.Println("Warning: The certificate will expire within 30 days.")
	} else {
		fmt.Println("The certificate is valid.")
	}
}
