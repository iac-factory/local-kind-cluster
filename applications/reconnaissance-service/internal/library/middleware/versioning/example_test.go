package versioning_test

import (
	"encoding/json"
	"net/http"

	"reconnaissance-service/internal/library/middleware"
	versioning2 "reconnaissance-service/internal/library/middleware/versioning"
)

func Example() {
	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().Version().Configuration(func(options *versioning2.Settings) { options.Version = versioning2.Version{Service: "0.0.0"} }).Middleware)

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
