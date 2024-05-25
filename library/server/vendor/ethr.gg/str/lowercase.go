package str

import (
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

	casing := cases.Lower(dialect, options...)

	return casing.String(v)
}
