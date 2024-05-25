package str

import (
	"golang.org/x/text/cases"
)

// Title will cast a string to title casing. If no options are provided,
// [cases.NoLower] will be appended by default.
func Title(v string, settings ...Variadic) string {
	var o Options
	for _, configuration := range settings {
		configuration(o)
	}

	options := o.Options
	if len(options) == 0 {
		options = append(options, cases.NoLower)
	}

	casing := cases.Title(dialect, options...)

	return casing.String(v)
}
