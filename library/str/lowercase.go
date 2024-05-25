package str

import (
	"log/slog"

	"golang.org/x/text/cases"
)

// Lowercase will cast a string to all-lower casing. If no options are provided,
// [cases.NoLower] will be appended by default.
func Lowercase(v string, settings ...Variadic) string {
	var o Options
	for _, configuration := range settings {
		configuration(o)
	}

	options := o.Options
	if len(options) == 0 {
		options = append(options, cases.NoLower)
	}

	if v == "" && o.Log {
		slog.Warn("Empty String Provided as Value")
	}

	casing := cases.Lower(dialect, options...)

	return casing.String(v)
}
