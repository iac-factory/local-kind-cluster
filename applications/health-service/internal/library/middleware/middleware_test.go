package middleware_test

import (
	"log/slog"
	"testing"

	"health-service/internal/library/middleware"
)

func Test(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug - 4)

	t.Run("Interface(s)", func(t *testing.T) {
		if v := middleware.New(); v == nil {
			t.Fatalf("Null Middleware Interface Returned")
		}

		t.Run("Envoy", func(t *testing.T) {
			if v := middleware.New().Envoy(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Path", func(t *testing.T) {
			if v := middleware.New().Path(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Server", func(t *testing.T) {
			if v := middleware.New().Server(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Service", func(t *testing.T) {
			if v := middleware.New().Service(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("State", func(t *testing.T) {
			if v := middleware.New().State(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Timeout", func(t *testing.T) {
			if v := middleware.New().Timeout(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Tracer", func(t *testing.T) {
			if v := middleware.New().Tracer(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Version", func(t *testing.T) {
			if v := middleware.New().Version(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})

		t.Run("Logs", func(t *testing.T) {
			if v := middleware.New().Logs(); v == nil {
				t.Fatalf("Null Middleware Interface Returned")
			}
		})
	})
}
