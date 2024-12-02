package expiration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"reconnaissance-service/internal/api"
	"reconnaissance-service/internal/library/middleware"

	"reconnaissance-service/internal/library/middleware/keystore"
)

func Test(t *testing.T) {
	// t.Skip("Test Refactoring In-Progress")

	ctx := context.Background()

	ctx = context.WithValue(ctx, keystore.Keys().Service(), "service")

	middlewares := middleware.Middleware()
	middlewares.Add(middleware.New().Telemetry().Middleware)

	mux := http.NewServeMux()

	api.Router(mux)

	handler := middlewares.Handler(mux)

	server := httptest.NewServer(handler)

	server.Config.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	defer server.Close()

	client := server.Client()

	t.Run("Expiration", func(t *testing.T) {
		t.Run("Unit-Test", func(t *testing.T) { // no proxy configuration
			t.Run("200", func(t *testing.T) {
				const hostname = "google.com"
				const port = 443

				var body bytes.Buffer
				json.NewEncoder(&body).Encode(map[string]interface{}{"hostname": hostname, "port": port})
				request, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/tls/expiration", server.URL), &body)
				if e != nil {
					t.Fatal(e)
				}

				response, exception := client.Do(request)
				if exception != nil {
					t.Fatal(exception)
				}

				defer response.Body.Close()

				t.Log("Successfully Made Server-Client Request")

				buffer, e := io.ReadAll(response.Body)
				if e != nil {
					t.Fatal("Unable to Read Response Body")
				}

				t.Logf("Output: %s", string(buffer))

				if response.StatusCode != http.StatusOK {
					t.Fatalf("Expected Status Code (%d), Received (%d)", http.StatusOK, response.StatusCode)
				}

				t.Logf("Successfully Called Endpoint")
			})

			t.Run("404", func(t *testing.T) {
				const hostname = "fail.test"
				const port = 443

				var body bytes.Buffer
				json.NewEncoder(&body).Encode(map[string]interface{}{"hostname": hostname, "port": port})
				request, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/tls/expiration", server.URL), &body)
				if e != nil {
					t.Fatal(e)
				}

				response, exception := client.Do(request)
				if exception != nil {
					t.Fatal(exception)
				}

				defer response.Body.Close()

				t.Log("Successfully Made Server-Client Request")

				buffer, e := io.ReadAll(response.Body)
				if e != nil {
					t.Fatal("Unable to Read Response Body")
				}

				t.Logf("Output: %s", string(buffer))

				if response.StatusCode != http.StatusNotFound {
					t.Fatalf("Expected Status Code (%d), Received (%d)", http.StatusNotFound, response.StatusCode)
				}

				t.Logf("Successfully Called Endpoint")
			})
		})
	})
}
