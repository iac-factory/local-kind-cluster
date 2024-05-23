package api

import (
	"context"
	"encoding/json"
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
		})
	})

	return r
}
