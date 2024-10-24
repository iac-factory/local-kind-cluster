package remove_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"authentication-service/internal/testing/proxy"

	"authentication-service/internal/middleware"

	"authentication-service/internal/middleware/keystore"

	"authentication-service/internal/api"
	"authentication-service/internal/database"
	"authentication-service/internal/token"
	"authentication-service/models/users"
)

func Test(t *testing.T) {
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

	session := proxy.New()

	session.Start(ctx)

	var record users.User
	var jwtstring string

	defer t.Cleanup(func() {
		connection, e := database.Connection(ctx)
		if e != nil {
			t.Fatal(e)
		}

		if e := users.New().Delete(ctx, connection, &users.DeleteParams{Email: record.Email, ID: record.ID}); e != nil {
			t.Fatal(e)
		}
	})

	t.Run("Setup", func(t *testing.T) {
		t.Helper()

		connection, e := database.Connection(ctx)
		if e != nil {
			t.Fatal(e)
		}

		r, e := users.New().Create(ctx, connection, &users.CreateParams{Email: "test-authentication-service-email-remove@x-ethr.gg", Password: "test-password"})
		if e != nil {
			t.Fatal(e)
		}

		record = r

		v, e := token.Create(ctx, record.Email)
		if e != nil {
			t.Fatal(e)
		}

		jwtstring = v
	})

	t.Run("Remove", func(t *testing.T) {
		request, e := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%d", server.URL, record.ID), nil)
		if e != nil {
			t.Fatal(e)
		}

		request.Header.Set("X-Testing-Authorization", fmt.Sprintf("Bearer %s", jwtstring))

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
	})
}
