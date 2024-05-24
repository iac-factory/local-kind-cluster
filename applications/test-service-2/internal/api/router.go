package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

// Router returns a chi.Router with bootstrapped middlewares.
func Router(service, version string) chi.Router {
	r := chi.NewRouter()

	logger := httplog.NewLogger(service, httplog.Options{
		LogLevel:         slog.LevelDebug,
		LevelFieldName:   slog.LevelKey,
		MessageFieldName: slog.MessageKey,
		Tags: map[string]string{
			"version": version,
		},
		JSON:               false,
		Concise:            false,
		RequestHeaders:     false,
		HideRequestHeaders: nil,
		ResponseHeaders:    false,
		QuietDownRoutes:    []string{},
		QuietDownPeriod:    60 * time.Second,
		TimeFieldFormat:    time.RFC822, // @todo evaluate on production or development time.RFC3339Nano,
		TimeFieldName:      slog.TimeKey,
		SourceFieldName:    "",
		Writer:             os.Stdout,
	})

	r.Use(middleware.Recoverer)
	r.Use(httplog.Handler(logger))

	r.Use(middleware.RequestID)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RealIP)
	r.Use(middleware.RedirectSlashes)

	var health = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "ok",
		}

		w.Header().Set("Content-Type", "Application/JSON")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)

		return
	})

	r.Handle("/health", health)

	r.Route("/{version}", func(r chi.Router) {
		r.Use(func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				version := chi.URLParam(r, "version")
				ctx := context.WithValue(r.Context(), "version", version)
				handler.ServeHTTP(w, r.WithContext(ctx))
			})
		})

		r.Route("/{service}", func(r chi.Router) {
			r.Use(func(handler http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					service := chi.URLParam(r, "service")
					ctx := context.WithValue(r.Context(), "service", service)
					handler.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			r.Handle("/health", health)

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				id := middleware.GetReqID(ctx)
				version := ctx.Value("version").(string)
				service := ctx.Value("service").(string)

				response := map[string]string{
					"route":   "root",
					"version": version,
					"service": service,
				}

				w.Header().Set("Content-Type", "Application/JSON")
				w.Header().Set("X-Request-ID", id)
				w.Header().Set("X-API-Service", service)
				w.Header().Set("X-API-Version", version)
				w.WriteHeader(http.StatusOK)

				json.NewEncoder(w).Encode(response)

				return
			})

			r.Get("/alpha-failure", func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				var client http.Client
				request, e := http.NewRequestWithContext(ctx, http.MethodGet, "http://test-service-1-a.development.svc.cluster.local:8080", nil)
				if e != nil {
					http.Error(w, "unable to create request", http.StatusInternalServerError)
					return
				}

				response, e := client.Do(request)
				if e != nil {
					slog.ErrorContext(ctx, "Error Making Internal Request", slog.String("error", e.Error()))
					http.Error(w, "error while making internal request", http.StatusInternalServerError)
					return
				}

				var structure map[string]interface{}
				content, e := io.ReadAll(response.Body)
				if e != nil {
					http.Error(w, "unable to read response body", http.StatusInternalServerError)
					return
				}

				slog.DebugContext(ctx, "Response", slog.String("raw", string(content)))

				if e := json.Unmarshal(content, &structure); e != nil {
					http.Error(w, "exception while unmarshalling response buffer", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
				w.Header().Set("X-Request-ID", response.Header.Get("X-Request-ID"))
				w.Header().Set("X-API-Service", response.Header.Get("X-API-Service"))
				w.Header().Set("X-API-Version", response.Header.Get("X-API-Version"))
				w.WriteHeader(response.StatusCode)

				json.NewEncoder(w).Encode(structure)

				return
			})
		})
	})

	return r
}
