package state_test

import (
	"encoding/json"
	"net/http"

	"user-service/internal/library/middleware"
)

func Example() {
	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().State().Middleware)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		version := middleware.New().Version().Value(ctx)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version)
	})

	http.ListenAndServe(":8080", middlewares.Handler(mux))
}
