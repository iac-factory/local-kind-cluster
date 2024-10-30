package telemetrics_test

import (
	"encoding/json"
	"net/http"

	"authentication-service/internal/library/middleware"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Telemetry().Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		telemetry := middleware.New().Telemetry().Value(ctx)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(telemetry)
	})

	http.ListenAndServe(":8080", handler)
}
