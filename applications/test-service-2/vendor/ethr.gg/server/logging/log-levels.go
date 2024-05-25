package logging

import "log/slog"

const (
	Trace = slog.LevelDebug - 4
	Debug = slog.LevelDebug
	Info  = slog.LevelInfo
	Warn  = slog.LevelWarn
	Error = slog.LevelError
)
