package versioning_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"verification-service/internal/library/middleware"
	versioning2 "verification-service/internal/library/middleware/versioning"
)

func Test(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug - 4)

	t.Run("Versioning-Middleware", func(t *testing.T) {
		middlewares := middleware.Middleware()
		middlewares.Add(middleware.New().Version().Configuration(func(options *versioning2.Settings) {
			options.Version = versioning2.Version{
				Service: "0.0.0",
			}
		}).Middleware)

		mux := http.NewServeMux()

		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			version := middleware.New().Version().Value(ctx)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(version)
		})

		server := httptest.NewServer(middlewares.Handler(mux))

		defer server.Close()

		t.Run("Successful-Version-Information", func(t *testing.T) {
			client := server.Client()
			request, e := http.NewRequest(http.MethodGet, server.URL, nil)
			if e != nil {
				t.Fatal(e)
			}

			response, exception := client.Do(request)
			if exception != nil {
				t.Fatal(exception)
			}

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("Unexpected Status Code: %d", response.StatusCode)
			}

			content, e := io.ReadAll(response.Body)
			if e != nil {
				t.Fatalf("Couldn't Read Response Body: %v", e)
			}

			t.Logf("Response: %s", string(content))
		})
	})
}
