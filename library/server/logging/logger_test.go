package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
)

func Test(t *testing.T) {
	Verbose(true)

	ctx := context.Background()

	t.Run("New", func(t *testing.T) {
		t.Run("Enabled-Trace-Log-Level", func(t *testing.T) {
			var w bytes.Buffer
			options := Options{Settings: &slog.HandlerOptions{AddSource: false, Level: slog.Level(0), ReplaceAttr: nil}}

			handler := Logger(&w, options)
			instance := slog.New(handler)
			slog.SetDefault(instance)

			slog.Log(ctx, slog.LevelInfo, "Test Message", slog.Group("wtf", slog.Group("partent09", slog.Group("parent", slog.Group(t.Name(), slog.String("value", "example"))))))
			slog.Log(ctx, slog.LevelInfo, "Test-Message-2", slog.String("key", "value"), slog.Bool("test", true), slog.Group("group", slog.Int64("int", int64(64))))
			if w.Len() == 0 {
				t.Fatalf("Invalid Log Message - Output Should be Enabled")
			}

			t.Logf("Success: %s, %s", t.Name(), w.String())
		})

		t.Run("Disabled-Trace-Log-Level", func(t *testing.T) {
			var w bytes.Buffer
			options := Options{Settings: &slog.HandlerOptions{AddSource: false, Level: slog.Level(0), ReplaceAttr: nil}}

			handler := Logger(&w, options)
			instance := slog.New(handler)
			slog.SetDefault(instance)

			slog.Log(ctx, slog.Level(-8), "Test Message", slog.Group(t.Name(), slog.String("value", "example")))

			if w.Len() > 0 {
				t.Fatalf("Invalid Log Message - Output Should be Disabled: %s", w.String())
			}

			t.Logf("Success: %s", t.Name())
		})
	})
}
