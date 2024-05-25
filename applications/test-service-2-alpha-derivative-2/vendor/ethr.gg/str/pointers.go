package str

import (
	"log/slog"
)

func Pointer(v string, settings ...Variadic) *string {
	var o Options
	for _, configuration := range settings {
		configuration(o)
	}

	if o.Log && v == "" {
		slog.Warn("Value is Empty String")
	}

	return &v
}

func Dereference(v *string, settings ...Variadic) string {
	var o Options
	for _, configuration := range settings {
		configuration(o)
	}

	if v == nil {
		if o.Log {
			slog.Warn("String nil Pointer - Returning Empty String")
		}

		return ""
	}

	return *(v)
}
