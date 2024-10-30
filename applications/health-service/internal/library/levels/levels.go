package levels

import (
	"log/slog"
)

// Exported constants representing [slog.Level].
//
// - Trace for tracing program's execution.
//
// - Debug for providing contextual information in debugging phase.
//
// - Info for informing about general system operations.
//
// - Notice for conditions that are not errors but might need handling.
//
// - Warning for warning conditions.
//
// - Error for error conditions.
//
// - Emergency for system-unusable conditions.
const (
	Trace     = slog.Level(-8)
	Debug     = slog.LevelDebug
	Info      = slog.LevelInfo
	Notice    = slog.Level(2)
	Warning   = slog.LevelWarn
	Error     = slog.LevelError
	Emergency = slog.Level(12)
)

// String returns the corresponding slog.Level value based on the provided level string.
// If the level string matches any of the predefined levels, the corresponding slog.Level value is returned.
// If no match is found, slog.LevelDebug is returned as the default value.
//
// Expected matches are:
//
//   - TRACE
//   - DEBUG
//   - INFO
//   - WARN
//   - ERROR
//   - EMERGENCY
func String(level string) (value slog.Level) {
	value = slog.LevelDebug

	switch level {
	case "TRACE":
		value = Trace
	case "DEBUG":
		value = Debug
	case "INFO":
		value = Info
	case "NOTICE":
		value = Notice
	case "WARN":
		value = Warning
	case "ERROR":
		value = Error
	case "EMERGENCY":
		value = Emergency
	}

	return
}
