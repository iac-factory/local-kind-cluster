package levels

import "log/slog"

// Attributes returns the updated attributes based on the provided groups and attribute.
//
//   - Handles updating the [slog.LevelKey] according to the [levels] package's additional custom log-levels: [Trace] and [Emergency]
//   - Usage is to be expected with [slog.HandlerOptions.ReplaceAttr].
func Attributes(groups []string, a slog.Attr) slog.Attr {
	if a.Key == (slog.LevelKey) {
		switch a.Value.Int64() {
		case int64(slog.Level(-8)):
			a.Value = slog.StringValue("TRACE")
		case int64(slog.Level(2)):
			a.Value = slog.StringValue("NOTICE")
		case int64(slog.Level(12)):
			a.Value = slog.StringValue("EMERGENCY")
		}
	}

	return a
}
