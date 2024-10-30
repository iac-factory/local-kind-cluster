package refresh_test

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

	"authentication-service/internal/library/testing/proxy"

	"authentication-service/internal/library/middleware"

	"authentication-service/internal/library/middleware/keystore"

	"authentication-service/internal/api"
	"authentication-service/internal/database"
	"authentication-service/models/users"
)

func Test(t *testing.T) {
	t.Skip("Test Refactoring In-Progress")

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

	const email, password = "test-login-user@x-ethr.gg", "test-password"

	session := proxy.New()

	session.Start(ctx)

	defer t.Cleanup(func() {
		session.Stop(ctx)

		connection, e := database.Connection(ctx)
		if e != nil {
			t.Fatalf("Unable to Connect to Database: %v", e)
		}

		if e := users.New().Clean(ctx, connection, email); e != nil {
			t.Errorf("Unable to Delete User: %v", e)
		}

		connection.Release()
	})

	var jwttoken string

	t.Run("Setup", func(t *testing.T) {
		t.Helper()

		t.Run("Registration", func(t *testing.T) {
			t.Helper()

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(map[string]interface{}{"email": email, "password": password})
			request, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/register", server.URL), &body)
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

			if response.StatusCode != http.StatusCreated {
				t.Fatalf("Expected Status Code (%d), Received (%d)", http.StatusCreated, response.StatusCode)
			}

			t.Logf("Successfully Registered User")

			jwttoken = string(buffer)
		})
	})

	t.Run("Refresh", func(t *testing.T) {
		t.Run("Invalid-Expiration-Timing", func(t *testing.T) {
			request, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/refresh", server.URL), nil)
			if e != nil {
				t.Fatal(e)
			}

			request.Header.Set("X-Testing-Authorization", fmt.Sprintf("Bearer %s", jwttoken))

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

			if response.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Expected Status Code (%d), Received (%d)", http.StatusTooManyRequests, response.StatusCode)
			}

			t.Logf("Successfully Received Too-Many-Requests Error")
		})

		t.Run("Unauthorized", func(t *testing.T) {
			request, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/refresh", server.URL), nil)
			if e != nil {
				t.Fatal(e)
			}

			request.Header.Set("X-Testing-Authorization", fmt.Sprintf("Bearer %s", "..."))

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

			if response.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected Status Code (%d), Received (%d)", http.StatusUnauthorized, response.StatusCode)
			}

			t.Logf("Successfully Received Unauthorized Error")
		})
	})
}
