package envoy_test

import (
	"encoding/json"
	"net/http"

	"reconnaissance-service/internal/library/middleware"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Envoy().Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := middleware.New().Envoy().Value(ctx)

		var response = map[string]interface{}{
			"value": value,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", handler)
}
