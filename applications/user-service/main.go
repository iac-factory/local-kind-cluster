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

	"user-service/internal/library/middleware"
	"user-service/internal/library/middleware/logs"
	"user-service/internal/library/middleware/name"
	"user-service/internal/library/middleware/servername"
	"user-service/internal/library/middleware/timeout"
	"user-service/internal/library/middleware/tracing"
	"user-service/internal/library/middleware/versioning"

	"user-service/internal/library/server"
	"user-service/internal/library/server/logging"
	"user-service/internal/library/server/telemetry"
	"user-service/internal/library/server/writer"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"

	"user-service/internal/api/avatar"
	"user-service/internal/api/me"
	"user-service/internal/api/registration"
	"user-service/internal/middleware/authentication"
)

// sname is a dynamically linked string value - defaults to "server" - which represents the server name.
var sname string = "server"

// service is a dynamically linked string value - defaults to "service" - which represents the service name.
var service string = "service"

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

// level represents the global log level - during initialization, the value is likely to change.
var level = slog.Level(-8)

func main() {
	// --> Middleware
	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().CORS().Middleware)
	middlewares.Add(middleware.New().Path().Middleware)
	middlewares.Add(middleware.New().Envoy().Middleware)
	middlewares.Add(middleware.New().RIP().Middleware)
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

		var response = map[string]interface{}{
			middleware.New().Service().Value(ctx): map[string]interface{}{
				"environment": environment,
				"path":        middleware.New().Path().Value(ctx),
				"service":     middleware.New().Service().Value(ctx),
				"api":         middleware.New().Version().Value(ctx).API,
				"version":     middleware.New().Version().Value(ctx).Service,
			},
		}

		// headers := telemetrics.New().Value(ctx).Headers
		//
		// {
		// 	// verification-service
		//
		// 	c := telemetry.Client(headers)
		//
		// 	namespace := os.Getenv("NAMESPACE")
		// 	if namespace == "" {
		// 		namespace = "development"
		// 	}
		//
		// 	url := fmt.Sprintf("http://verification-service.%s.svc.cluster.local:8080", namespace)
		//
		// 	request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	svc, e := c.Do(request)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	defer svc.Body.Close()
		//
		// 	content, e := io.ReadAll(svc.Body)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	// --> only successful responses will be in json format
		//
		// 	switch svc.StatusCode {
		// 	case http.StatusOK:
		// 		var mapping map[string]interface{}
		// 		if e := json.Unmarshal(content, &mapping); e != nil {
		// 			slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("error", e.Error()))
		// 			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 			return
		// 		}
		//
		// 		maps.Copy(response[service].(map[string]interface{}), mapping)
		// 	default: // note an error response is not returned
		// 		slog.ErrorContext(ctx, "Service Returned an Error", slog.String("url", url), slog.Int("status", svc.StatusCode), slog.String("response", string(content)))
		// 	}
		// }

		// {
		// 	// customer-service
		//
		// 	c := telemetry.Client(headers)
		//
		// 	namespace := os.Getenv("NAMESPACE")
		// 	if namespace == "" {
		// 		namespace = "development"
		// 	}
		//
		// 	url := fmt.Sprintf("http://customer-service.%s.svc.cluster.local:8080", namespace)
		//
		// 	request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	svc, e := c.Do(request)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	defer svc.Body.Close()
		//
		// 	content, e := io.ReadAll(svc.Body)
		// 	if e != nil {
		// 		slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		//
		// 	// --> only successful responses will be in json format
		//
		// 	switch svc.StatusCode {
		// 	case http.StatusOK:
		// 		var mapping map[string]interface{}
		// 		if e := json.Unmarshal(content, &mapping); e != nil {
		// 			slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("error", e.Error()))
		// 			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 			return
		// 		}
		//
		// 		maps.Copy(response[service].(map[string]interface{}), mapping)
		// 	default: // note an error response is not returned
		// 		slog.ErrorContext(ctx, "Service Returned an Error", slog.String("url", url), slog.Int("status", svc.StatusCode), slog.String("response", string(content)))
		// 	}
		// }

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		return
	})))

	mux.Handle("POST /register", otelhttp.WithRouteTag("/register", registration.Handler))
	mux.Handle("PATCH /avatar", otelhttp.WithRouteTag("/avatar", avatar.Patch))

	mux.Handle("GET /@me", otelhttp.WithRouteTag("/@me", authentication.Middleware(me.Handler)))

	// --> Start the HTTP server
	slog.Info("Starting Server ...", slog.String("local", fmt.Sprintf("http://localhost:%s", *(port))))

	handler := writer.Handle(middlewares.Handler(mux))
	handler = otelhttp.NewHandler(handler, "server", otelhttp.WithServerName(service), otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents))

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

	if os.Getenv("CI") == "true" {
		level = slog.LevelDebug
	}

	logging.Level(level)
	slog.SetLogLoggerLevel(level)
	if service == "service" && os.Getenv("CI") != "true" {
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

	handler := logging.Logger(func(o *logging.Options) { o.Service = service })
	logger = slog.New(handler)
	slog.SetDefault(logger)
}