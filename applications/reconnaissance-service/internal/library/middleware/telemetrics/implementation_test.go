package telemetrics_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"reconnaissance-service/internal/library/middleware"
)

func Test(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug - 4)

	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Telemetry().Middleware)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		telemetry := middleware.New().Telemetry().Value(ctx)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(telemetry)
	})

	server := httptest.NewServer(middlewares.Handler(mux))

	server.Config.BaseContext = func(net.Listener) context.Context {
		ctx := context.Background()

		return ctx
	}

	defer server.Close()

	t.Run("Versioning-Middleware", func(t *testing.T) {
		t.Run("Successful-Telemetry-Information", func(t *testing.T) {
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
