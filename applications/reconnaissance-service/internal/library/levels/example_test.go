package levels_test

import (
	"bytes"
	"context"
	"log/slog"

	"reconnaissance-service/internal/library/levels"
)

func Example() {
	ctx := context.Background()
	level := levels.Trace
	slog.SetLogLoggerLevel(level)

	options := &slog.HandlerOptions{
		Level: level,
	}

	var output bytes.Buffer

	slog.SetDefault(slog.New(slog.NewTextHandler(&output, options)))

	slog.Log(ctx, levels.Trace, "Example Trace Log Message")
}
