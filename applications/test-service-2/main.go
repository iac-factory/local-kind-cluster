package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"ethr.gg/headers"
	"ethr.gg/server"
	"ethr.gg/server/logging"
	"ethr.gg/server/middleware"
)

// service is a dynamically linked string value - defaults to "service" - which represents the service name.
var service string = "service"

// version is a dynamically linked string value - defaults to "development" - which represents the service's version.
var version string = "development" // production builds have version dynamically linked

// ctx, cancel represent the server's runtime context and cancellation handler.
var ctx, cancel = context.WithCancel(context.Background())

// port represents a cli flag that sets the server listening port
var port = flag.String("port", "8080", "Server Listening Port.")

func main() {
	mux := server.Mux()

	mux.HandleFunc("GET /{version}/{service}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		headers.Log(r, headers.Incoming)

		var payload = map[string]interface{}{
			middleware.New().Service().Value(ctx): map[string]interface{}{
				"service":     middleware.New().Service().Value(ctx),
				"version":     middleware.New().Version().Value(ctx).Service,
				"api-version": middleware.New().Version().Value(ctx).API,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(payload)
	})

	mux.HandleFunc("GET /{version}/{service}/alpha", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		headers.Log(r, headers.Incoming)

		client := http.Client{}
		request, e := http.NewRequestWithContext(ctx, "GET", "http://test-service-2-alpha.development.svc.cluster.local:8080", nil)
		if e != nil {
			slog.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
			http.Error(w, "error while generating GET request to internal service", http.StatusInternalServerError)
			return
		}

		headers.Propagate(r, request)

		response, e := client.Do(request)
		if e != nil {
			slog.ErrorContext(ctx, "Error Generating Response from Internal Service", slog.String("error", e.Error()))
			http.Error(w, "exception generating response from internal service", http.StatusInternalServerError)
			return
		}

		defer response.Body.Close()

		var structure interface{}
		if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
			http.Error(w, "unable to decode response body to normalized data structure", http.StatusInternalServerError)
			return
		}

		var payload = map[string]interface{}{
			middleware.New().Service().Value(ctx): map[string]interface{}{
				"service":     middleware.New().Service().Value(ctx),
				"version":     middleware.New().Version().Value(ctx).Service,
				"api-version": middleware.New().Version().Value(ctx).API,
				"response":    structure,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(payload)
	})

	api := server.Server(ctx, mux, service, version, *port)

	// Issue Cancellation Handler
	server.Interrupt(ctx, cancel, api)

	shutdown, e := server.Setup(ctx, service, version)
	if e != nil {
		panic(e)
	}

	defer func() {
		e = errors.Join(e, shutdown(context.Background()))
	}()

	// Start HTTP Server
	slog.InfoContext(ctx, "Starting Server ...", slog.String("port", *(port)), slog.String("service", service), slog.String("version", version))

	fmt.Print("\n")

	// <-- Blocking
	if e := api.ListenAndServe(); e != nil && !(errors.Is(e, http.ErrServerClosed)) {
		slog.ErrorContext(ctx, "Error During Server's Listen & Serve Call ...", slog.String("error", e.Error()))

		os.Exit(100)
	}

	// --> Exit
	{
		slog.InfoContext(ctx, "Graceful Shutdown Complete")

		// Waiter
		<-ctx.Done()
	}
}

func init() {
	flag.Parse()

	level := slog.Level(-8)
	if os.Getenv("CI") == "true" {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level.Level())

	if service == "service" && os.Getenv("CI") != "true" {
		_, file, _, ok := runtime.Caller(0)
		if ok {
			service = filepath.Base(filepath.Dir(file))
		}
	}

	options := logging.Options{Service: service, Settings: &slog.HandlerOptions{Level: level}}
	handler := logging.Logger(os.Stdout, options)

	slog.SetDefault(slog.New(handler))
}
