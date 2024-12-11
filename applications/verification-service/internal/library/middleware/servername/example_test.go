package servername_test

import (
	"encoding/json"
	"net/http"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/servername"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Server().Configuration(func(options *servername.Settings) { options.Server = "Server-Name-Value" }).Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := middleware.New().Server().Value(ctx)

		var response = map[string]interface{}{
			"value": value,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", handler)
}
