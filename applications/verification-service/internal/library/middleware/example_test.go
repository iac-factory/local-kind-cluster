package middleware_test

import (
	"encoding/json"
	"net/http"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/name"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = "example-service-name" }).Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := middleware.New().Service().Value(ctx)

		var response = map[string]interface{}{
			"value": value,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", handler)
}
