package cors_test

import (
	"encoding/json"
	"net/http"

	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/cors"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().CORS().Configuration(func(options *cors.Settings) { options.Debug = true }).Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := middleware.New().CORS().Value(ctx)

		var response = map[string]interface{}{
			"value": value,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", handler)
}
