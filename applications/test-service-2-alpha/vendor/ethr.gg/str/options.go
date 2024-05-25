package str

import "golang.org/x/text/cases"

type Options struct {
	// Log represents an optional flag that will log when potential, unexpected behavior could occur. E.g.
	// when using the Dereference function, log a warning that the pointer was nil.
	Log bool

	// Options represents an array of cases.Option. These are only applicable to certain casing functions.
	Options []cases.Option
}

// Variadic represents a functional constructor for the Options type.
type Variadic func(o Options)
