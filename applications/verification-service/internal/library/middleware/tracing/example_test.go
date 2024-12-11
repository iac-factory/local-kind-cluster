package tracing_test

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/tracing"
)

func Example() {
	var tracer = otel.Tracer("test-service")

	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().Tracer().Configuration(func(options *tracing.Settings) { options.Tracer = tracer }).Middleware)

	middlewares.Add(middleware.New().State().Middleware)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := middleware.New().Tracer().Value(ctx)
		ctx, span := tracer.Start(ctx, "example-handler", trace.WithAttributes(attribute.String("handler", "example-handler")))

		defer span.End()

		span.SetStatus(http.StatusOK, "N/A")

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	http.ListenAndServe(":8080", middlewares.Handler(mux))
}
