package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"test-service-1/internal/api"
)

// Server Runtime Context
var ctx, cancel = context.WithCancel(context.Background())

var (
	// hostname - Host-System's Hostname
	hostname string

	// exception - Server Start-Up Error
	exception error

	// server - the HTTP API Server
	server *http.Server

	//go:embed global-bundle.pem
	bundle embed.FS // --> ignored; used to compile binary with pem file
)

var service string = "service"     // production builds have service dynamically linked
var version string = "development" // production builds have version dynamically linked

var (
	port = flag.String("port", "8080", "Server Listening Port.")
)

func main() {
	// Issue Cancellation Handler
	api.Interrupt(ctx, cancel, server)

	shutdown, e := setupOTelSDK(ctx)
	if e != nil {
		panic(e)
	}

	defer func() {
		e = errors.Join(e, shutdown(context.Background()))
	}()

	// Start HTTP Server
	slog.InfoContext(ctx, "Starting Server ...", slog.String("hostname", hostname), slog.String("port", *(port)), slog.String("service", service), slog.String("version", version))

	fmt.Print("\n")

	// <-- Blocking
	if e := server.ListenAndServe(); e != nil && !(errors.Is(e, http.ErrServerClosed)) {
		slog.ErrorContext(ctx, "Error During Server's Listen & Serve Call ...", slog.String("error", e.Error()))

		os.Exit(100)
	}

	// --> Exit
	{
		slog.Log(ctx, slog.LevelInfo, "Graceful Shutdown Complete")

		// Waiter
		<-ctx.Done()
	}
}

func init() {
	flag.Parse()

	slog.SetLogLoggerLevel(slog.LevelDebug)

	if buffer := strings.TrimSpace(version); len(buffer) > 0 {
		if e := os.Setenv("VERSION", string(buffer)); e != nil {
			slog.WarnContext(ctx, "Unable to Set VERSION Environment Variable", slog.String("error", e.Error()))
		}
	}

	if hostname, exception = os.Hostname(); exception != nil {
		slog.ErrorContext(ctx, "Unable to Register Hostname", slog.String("error", exception.Error()))

		cancel()

		os.Exit(100)
	}

	server = api.Server(ctx, api.Router(service, version), *port)
}
