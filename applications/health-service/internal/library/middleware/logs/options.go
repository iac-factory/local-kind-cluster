package logs

import (
	"log/slog"

	"health-service/internal/library/middleware/types"
)

type Settings struct {
	Logger *slog.Logger
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{
		Logger: slog.Default(),
	}
}