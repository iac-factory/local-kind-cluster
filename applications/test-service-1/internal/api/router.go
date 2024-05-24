package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Router returns a chi.Router with bootstrapped middlewares.
func Router(service, version string) *http.ServeMux {
	mux := http.NewServeMux()

	// logger := httplog.NewLogger(service, httplog.Options{
	// 	LogLevel:         slog.LevelDebug,
	// 	LevelFieldName:   slog.LevelKey,
	// 	MessageFieldName: slog.MessageKey,
	// 	Tags: map[string]string{
	// 		"version": version,
	// 	},
	// 	JSON:               false,
	// 	Concise:            false,
	// 	RequestHeaders:     false,
	// 	HideRequestHeaders: nil,
	// 	ResponseHeaders:    false,
	// 	QuietDownRoutes:    []string{},
	// 	QuietDownPeriod:    60 * time.Second,
	// 	TimeFieldFormat:    time.RFC822, // @todo evaluate on production or development time.RFC3339Nano,
	// 	TimeFieldName:      slog.TimeKey,
	// 	SourceFieldName:    "",
	// 	Writer:             os.Stdout,
	// })

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handler := func(pattern string, h func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handle := otelhttp.WithRouteTag(pattern, http.HandlerFunc(h))
		mux.Handle(pattern, handle)
	}

	var health = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "ok",
		}

		w.Header().Set("Content-Type", "Application/JSON")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)

		return
	})

	handler("GET /health", health)

	var propagate = func(source, target *http.Request) {
		headers := []string{
			http.CanonicalHeaderKey("portal"),
			http.CanonicalHeaderKey("device"),
			http.CanonicalHeaderKey("user"),
			http.CanonicalHeaderKey("travel"),
			http.CanonicalHeaderKey("x-request-id"),
			http.CanonicalHeaderKey("x-b3-traceid"),
			http.CanonicalHeaderKey("x-b3-spanid"),
			http.CanonicalHeaderKey("x-b3-parentspanid"),
			http.CanonicalHeaderKey("x-b3-sampled"),
			http.CanonicalHeaderKey("x-b3-flags"),
			http.CanonicalHeaderKey("x-ot-span-context"),
		}

		for key := range headers {
			header := headers[key]

			target.Header.Add(header, source.Header.Get(header))
		}
	}

	handler("GET /alpha", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slog.DebugContext(ctx, "Total Headers", slog.Int("total", len(r.Header)))
		for header, values := range r.Header {
			slog.DebugContext(ctx, "Headers", slog.String("header", header), slog.String("value", values[0]))
		}

		slog.DebugContext(ctx, "Host", slog.String("value", r.Host))
		slog.DebugContext(ctx, "Referrer", slog.String("value", r.Referer()))

		var client http.Client
		request, e := http.NewRequestWithContext(ctx, http.MethodGet, "http://test-service-1-a.development.svc.cluster.local:8080", nil)
		if e != nil {
			http.Error(w, "unable to create request", http.StatusInternalServerError)
			return
		}

		propagate(r, request)

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

	handler("GET /bravo", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var client http.Client
		request, e := http.NewRequestWithContext(ctx, http.MethodGet, "http://test-service-1-b.development.svc.cluster.local:8080", nil)
		if e != nil {
			http.Error(w, "unable to create request", http.StatusInternalServerError)
			return
		}

		propagate(r, request)

		response, e := client.Do(request)
		if e != nil {
			slog.ErrorContext(ctx, "Error Making Internal Request", slog.String("error", e.Error()))
			http.Error(w, "error while making internal request", http.StatusInternalServerError)
			return
		}

		defer response.Body.Close()

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

	// r.Route("/{version}", func(r chi.Router) {
	// 	r.Use(func(handler http.Handler) http.Handler {
	// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			version := chi.URLParam(r, "version")
	// 			ctx := context.WithValue(r.Context(), "version", version)
	// 			handler.ServeHTTP(w, r.WithContext(ctx))
	// 		})
	// 	})
	//
	// 	r.Route("/{service}", func(r chi.Router) {
	// 		r.Use(func(handler http.Handler) http.Handler {
	// 			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 				service := chi.URLParam(r, "service")
	// 				ctx := context.WithValue(r.Context(), "service", service)
	// 				handler.ServeHTTP(w, r.WithContext(ctx))
	// 			})
	// 		})
	//
	// 		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 			ctx := r.Context()
	//
	// 			version := ctx.Value("version").(string)
	// 			service := ctx.Value("service").(string)
	//
	// 			response := map[string]string{
	// 				"route":   "root",
	// 				"version": version,
	// 				"service": service,
	// 			}
	//
	// 			w.Header().Set("Content-Type", "Application/JSON")
	// 			w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
	// 			w.Header().Set("X-API-Service", service)
	// 			w.Header().Set("X-API-Version", version)
	// 			w.WriteHeader(http.StatusOK)
	//
	// 			json.NewEncoder(w).Encode(response)
	//
	// 			return
	// 		})
	//
	// 		r.Get("/alpha", func(w http.ResponseWriter, r *http.Request) {
	// 			ctx := r.Context()
	//
	// 			slog.DebugContext(ctx, "Total Headers", slog.Int("total", len(r.Header)))
	// 			for header, values := range r.Header {
	// 				slog.DebugContext(ctx, "Headers", slog.String("header", header), slog.String("value", values[0]))
	// 			}
	//
	// 			slog.DebugContext(ctx, "Host", slog.String("value", r.Host))
	// 			slog.DebugContext(ctx, "Referrer", slog.String("value", r.Referer()))
	//
	// 			var client http.Client
	// 			request, e := http.NewRequestWithContext(ctx, http.MethodGet, "http://test-service-1-a.development.svc.cluster.local:8080", nil)
	// 			if e != nil {
	// 				http.Error(w, "unable to create request", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			propagate(r, request)
	//
	// 			response, e := client.Do(request)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Error Making Internal Request", slog.String("error", e.Error()))
	// 				http.Error(w, "error while making internal request", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			var structure map[string]interface{}
	// 			content, e := io.ReadAll(response.Body)
	// 			if e != nil {
	// 				http.Error(w, "unable to read response body", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			slog.DebugContext(ctx, "Response", slog.String("raw", string(content)))
	//
	// 			if e := json.Unmarshal(content, &structure); e != nil {
	// 				http.Error(w, "exception while unmarshalling response buffer", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
	// 			w.Header().Set("X-Request-ID", response.Header.Get("X-Request-ID"))
	// 			w.Header().Set("X-API-Service", response.Header.Get("X-API-Service"))
	// 			w.Header().Set("X-API-Version", response.Header.Get("X-API-Version"))
	// 			w.WriteHeader(response.StatusCode)
	//
	// 			json.NewEncoder(w).Encode(structure)
	//
	// 			return
	// 		})
	//
	// 		r.Get("/bravo", func(w http.ResponseWriter, r *http.Request) {
	// 			ctx := r.Context()
	//
	// 			var client http.Client
	// 			request, e := http.NewRequestWithContext(ctx, http.MethodGet, "http://test-service-1-b.development.svc.cluster.local:8080", nil)
	// 			if e != nil {
	// 				http.Error(w, "unable to create request", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			propagate(r, request)
	//
	// 			response, e := client.Do(request)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Error Making Internal Request", slog.String("error", e.Error()))
	// 				http.Error(w, "error while making internal request", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			defer response.Body.Close()
	//
	// 			var structure map[string]interface{}
	//
	// 			content, e := io.ReadAll(response.Body)
	// 			if e != nil {
	// 				http.Error(w, "unable to read response body", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			slog.DebugContext(ctx, "Response", slog.String("raw", string(content)))
	//
	// 			if e := json.Unmarshal(content, &structure); e != nil {
	// 				http.Error(w, "exception while unmarshalling response buffer", http.StatusInternalServerError)
	// 				return
	// 			}
	//
	// 			w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
	// 			w.Header().Set("X-Request-ID", response.Header.Get("X-Request-ID"))
	// 			w.Header().Set("X-API-Service", response.Header.Get("X-API-Service"))
	// 			w.Header().Set("X-API-Version", response.Header.Get("X-API-Version"))
	// 			w.WriteHeader(response.StatusCode)
	//
	// 			json.NewEncoder(w).Encode(structure)
	//
	// 			return
	// 		})
	// 	})
	// })

	return mux
}
