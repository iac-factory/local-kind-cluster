package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"user-service/internal/api"
	"user-service/internal/library/middleware"
	"user-service/internal/library/middleware/logs"
	"user-service/internal/library/middleware/name"
	"user-service/internal/library/middleware/servername"
	"user-service/internal/library/middleware/telemetrics"
	"user-service/internal/library/middleware/timeout"
	"user-service/internal/library/middleware/tracing"
	"user-service/internal/library/middleware/versioning"

	"user-service/internal/library/server"
	"user-service/internal/library/server/logging"
	"user-service/internal/library/server/telemetry"
	"user-service/internal/library/server/writer"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
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

	mux.Handle("GET /", otelhttp.WithRouteTag("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const name = "metadata"

		ctx := r.Context()
		ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

		labeler, _ := otelhttp.LabelerFromContext(ctx)

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

		// --> channel response structure
		type channels struct {
			authentication chan map[string]interface{}
			verification   chan map[string]interface{}
		}

		// --> telemetry-capable client + header(s)
		c := telemetry.Client(telemetrics.New().Value(ctx).Headers)

		// --> add service source header to avoid recursive callback(s)
		maps.Copy(c.Headers, map[string]string{"X-Service-Source": service})

		// --> establish an error-group for external response-handling
		g := new(errgroup.Group)

		// --> external service response(s)
		var responses = channels{
			authentication: make(chan map[string]interface{}, 1),
			verification:   make(chan map[string]interface{}, 1),
		}

		// --> closures for all structure's channel(s)
		defer close(responses.authentication)
		defer close(responses.verification)

		g.Go(func() error { // authentication-service
			if strings.Contains(strings.ToLower(r.Header.Get("X-Service-Source")), "authentication") {
				slog.DebugContext(ctx, "Skipping Recursive Authentication-Service Callback")
				responses.authentication <- map[string]interface{}{}

				return nil
			}

			u := fmt.Sprintf("%s://%s:%d", "http", "authentication-service", 8080)
			request, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if e != nil {
				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Generate Request"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			svc, e := c.Do(request)
			if e != nil {
				var lookup *net.DNSError
				if errors.As(e, &lookup) && os.Getenv("CI") == "" && lookup.IsNotFound { // local development - no such host
					slog.WarnContext(ctx, "URL Isn't Available in Local Environment", slog.String("error", lookup.Error()))
					responses.authentication <- map[string]interface{}{
						"authentication-service": map[string]interface{}{
							"status": "unavailable",
						},
					}

					return nil
				}

				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Process HTTP-Response"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			defer svc.Body.Close()

			content, e := io.ReadAll(svc.Body)
			if e != nil {
				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Read Raw Response"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			switch svc.StatusCode {
			case http.StatusOK: // --> only successful responses will be in json format
				var mapping map[string]interface{}
				if e := json.Unmarshal(content, &mapping); e != nil {
					exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Unmarshal Response"}}
					slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
					return &exception
				}

				responses.authentication <- mapping
			default: // note an error response is not returned
				e = fmt.Errorf("unexpected status code: %d, %s", svc.StatusCode, string(content))
				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unexpected HTTP Status Code"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			return nil
		})

		g.Go(func() error { // verification-service
			if strings.Contains(strings.ToLower(r.Header.Get("X-Service-Source")), "verification") {
				slog.DebugContext(ctx, "Skipping Recursive Verification-Service Callback")
				responses.authentication <- map[string]interface{}{}

				return nil
			}

			u := fmt.Sprintf("%s://%s:%d", "http", "verification-service", 8080)
			request, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if e != nil {
				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Generate Request"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			svc, e := c.Do(request)
			if e != nil {
				var lookup *net.DNSError
				if errors.As(e, &lookup) && os.Getenv("CI") == "" && lookup.IsNotFound { // local development - no such host
					slog.WarnContext(ctx, "URL Isn't Available in Local Environment", slog.String("error", lookup.Error()))
					responses.verification <- map[string]interface{}{
						"verification-service": map[string]interface{}{
							"status": "unavailable",
						},
					}

					return nil
				}

				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Process HTTP-Response"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			defer svc.Body.Close()

			content, e := io.ReadAll(svc.Body)
			if e != nil {
				logger.WarnContext(ctx, reflect.TypeOf(e).String())

				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Read Raw Response"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			switch svc.StatusCode {
			case http.StatusOK: // --> only successful responses will be in json format
				var mapping map[string]interface{}
				if e := json.Unmarshal(content, &mapping); e != nil {
					exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unable to Unmarshal Response"}}
					slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
					return &exception
				}

				responses.verification <- mapping
			default: // note an error response is not returned
				e = fmt.Errorf("unexpected status code: %d, %s", svc.StatusCode, string(content))
				exception := server.Exception{Code: http.StatusInternalServerError, Internal: &server.Internal{Error: e, Message: "Unexpected HTTP Status Code"}}
				slog.ErrorContext(ctx, exception.Internal.Message, slog.String("error", e.Error()))
				return &exception
			}

			return nil
		})

		// --> an error is only returned upon a fatal, internal server error
		if e := g.Wait(); e != nil {
			var exception *server.Exception
			if !(errors.As(e, &exception)) {
				slog.ErrorContext(ctx, "A Fatal, Unexpected Internal Server Error Has Occurred", slog.String("error", e.Error()))

				labeler.Add(attribute.Bool("error", true))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			slog.ErrorContext(ctx, "Server Exception Occurred", slog.String("error", e.Error()), slog.String("status", exception.Status), slog.Int("code", exception.Code), slog.Group("internal", slog.String("message", exception.Internal.Message), slog.Any("error", exception.Internal.Error)))
			exception.Response(w)
			return
		}

		slog.DebugContext(ctx, "Completed External Request(s)")

		maps.Copy(response[service].(map[string]interface{}), <-responses.authentication)
		maps.Copy(response[service].(map[string]interface{}), <-responses.verification)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		return
	})))

	api.Router(mux)

	// --> Start the HTTP server
	slog.Info("Starting Server ...", slog.String("port", *(port)))

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
