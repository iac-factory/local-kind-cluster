package servername_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"health-service/internal/library/middleware"
	"health-service/internal/library/middleware/servername"
)

func Test(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug - 4)

	t.Run("Server-Name-Middleware", func(t *testing.T) {
		middlewares := middleware.Middleware()
		middlewares.Add(middleware.New().Server().Configuration(func(options *servername.Settings) { options.Server = "Test-Server-Name" }).Middleware)

		mux := http.NewServeMux()

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

		server := httptest.NewServer(middlewares.Handler(mux))

		defer server.Close()

		t.Run("Successful", func(t *testing.T) {
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

			t.Logf("Headers: %v", response.Header)
		})
	})
}
