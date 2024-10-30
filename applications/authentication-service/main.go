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
	"time"

	"authentication-service/internal/library/server"
	"authentication-service/internal/library/server/logging"
	"authentication-service/internal/library/server/telemetry"
	"authentication-service/internal/library/server/writer"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"

	"authentication-service/internal/library/middleware/logs"
	"authentication-service/internal/library/middleware/name"
	"authentication-service/internal/library/middleware/servername"
	"authentication-service/internal/library/middleware/timeout"
	"authentication-service/internal/library/middleware/tracing"
	"authentication-service/internal/library/middleware/versioning"

	"authentication-service/internal/library/middleware"

	"authentication-service/internal/api"
)

// sname is a dynamically linked string value - defaults to "local-http-server" - which represents the server name.
var sname string = "local-http-server"

// service is a dynamically linked string value - defaults to "" - which represents the service name.
var service string

// version is a dynamically linked string value - defaults to "latest" - which represents the service's build version.
var version string = "latest"

// environment is dynamically updated according to "ENVIRONMENT" environment variable.
var environment string = "local"

// ctx, cancel represent the server's runtime context and cancellation handler.
var ctx, cancel = context.WithCancel(context.Background())

// port represents a cli flag that sets the server listening port.
var port = flag.String("port", "8080", "Server Listening Port.")

// tracer is the runtime's [otel.Tracer]. Used in the main function's middleware.
var tracer = otel.Tracer(service)

// logger represents an [slog.Logger] interface -- hydrated during the init call and then used in middleware found in main.
var logger *slog.Logger

// level represents the runtime's default log level. note that this variable is evaluated and changed according to a wide variety of factors.
var level = logging.Global()

func main() {
	// --> Middleware
	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().CORS().Middleware)
	middlewares.Add(middleware.New().Path().Middleware)
	middlewares.Add(middleware.New().Envoy().Middleware)
	middlewares.Add(middleware.New().Telemetry().Middleware)
	middlewares.Add(middleware.New().Timeout().Configuration(func(options *timeout.Settings) { options.Timeout = 30 * time.Second }).Middleware)
	middlewares.Add(middleware.New().Server().Configuration(func(options *servername.Settings) { options.Server = sname }).Middleware)
	middlewares.Add(middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = service }).Middleware)
	middlewares.Add(middleware.New().Version().Configuration(func(options *versioning.Settings) { options.Version.Service = version }).Middleware)
	middlewares.Add(middleware.New().Tracer().Configuration(func(options *tracing.Settings) { options.Tracer = tracer }).Middleware)
	middlewares.Add(middleware.New().Logs().Configuration(func(options *logs.Settings) { options.Logger = logger }).Middleware)

	// --> HTTP Handler(s)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.Health)

	mux.Handle("GET /", otelhttp.WithRouteTag("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const name = "metadata"

		ctx := r.Context()
		ctx, span := middleware.New().Tracer().Value(ctx).Start(ctx, name)

		defer span.End()

		instance := middleware.New()
		var response = map[string]interface{}{
			middleware.New().Service().Value(ctx): map[string]interface{}{
				"environment": environment,
				"path":        instance.Path().Value(ctx),
				"service":     instance.Service().Value(ctx),
				"api":         instance.Version().Value(ctx).API,
				"version":     instance.Version().Value(ctx).Service,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		return
	})))
	// mux.Handle("POST /login", otelhttp.WithRouteTag("/login", login.Handler))
	//
	// mux.Handle("GET /logout", otelhttp.WithRouteTag("/logout", logout.Handler))
	//
	// mux.Handle("POST /register", otelhttp.WithRouteTag("/register", registration.Handler))
	//
	// mux.Handle("POST /refresh", otelhttp.WithRouteTag("/refresh", authentication.Middleware(refresh.Handler)))
	//
	// mux.Handle("POST /", otelhttp.WithRouteTag("/", authentication.Middleware(create.Handler)))
	// mux.Handle("DELETE /", otelhttp.WithRouteTag("/", authentication.Middleware(deletion.Handler)))

	api.Router(mux)

	// --> Start the HTTP server
	slog.Info("Starting Server ...", slog.String("local", fmt.Sprintf("http://localhost:%s", *(port))))

	handler := writer.Handle(middlewares.Handler(mux))
	handler = otelhttp.NewHandler(handler, "server", otelhttp.WithServerName(service))
	otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents)

	api := server.Server(ctx, handler, *port)

	// --> Issue Cancellation Handler
	server.Interrupt(ctx, cancel, api)

	// --> Telemetry Setup + Cancellation Handler
	shutdown, e := telemetry.Setup(ctx, service, version, func(options *telemetry.Settings) {
		if version == "development" && os.Getenv("CI") == "" {
			options.Zipkin.Enabled = false

			options.Tracer.Local = true
			options.Metrics.Local = true
			options.Logs.Local = true
		}
	})

	if e != nil {
		panic(e)
	}

	defer func() {
		e = errors.Join(e, shutdown(ctx))
	}()

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

	if service == "" && os.Getenv("CI") != "true" {
		_, file, _, ok := runtime.Caller(0)
		if ok {
			service = filepath.Base(filepath.Dir(file))
		}

		if e := os.Setenv("SERVICE", service); e != nil {
			slog.ErrorContext(ctx, "Unable to Set SERVICE Environment Variable", slog.String("error", e.Error()))
			panic(e)
		}
	}

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		environment = v
	}

	defer os.Setenv("ENVIRONMENT", environment)

	handler := logging.Logger(func(o *logging.Options) { o.Service = service })

	logger = slog.New(handler)

	slog.SetDefault(logger)
}
