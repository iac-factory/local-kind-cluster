package logs_test

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"reconnaissance-service/internal/library/middleware"
)

func Example() {
	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Logs().Middleware)

	mux := http.NewServeMux()

	handler := middlewares.Handler(mux)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger := middleware.New().Logs().Value(ctx)

		var response = map[string]interface{}{
			"key": "value",
		}

		logger = logger.With(slog.String("key", "value"))

		logger.InfoContext(ctx, "Response", slog.Any("response", response))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", handler)
}
