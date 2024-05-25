package color

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/term"
)

type ANSI string

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	italic = "\033[3m"
	black  = "\033[30m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

// var (
//	ci bool // ci represents a false isTTY
// )

// Black applies the black color to the input string(s) and returns it as an ANSI string.
// It uses the Black() function from the ANSI package to convert the input to black color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Black(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Black(input...)))

	*c = ANSI(v)

	return c
}

// Red applies the red color to the input string(s) and returns it as an ANSI string.
// It uses the Red() function from the ANSI package to convert the input to red color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Red(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Red(input...)))

	*c = ANSI(v)

	return c
}

// Green applies the green color to the input string(s) and returns it as an ANSI string.
// It uses the Green() function from the ANSI package to convert the input to green color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Green(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Green(input...)))

	*c = ANSI(v)

	return c
}

// Yellow applies the yellow color to the input string(s) and returns it as an ANSI string.
// It uses the Yellow() function from the ANSI package to convert the input to yellow color.
// If the current operating system is not Windows and the ci variable is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Yellow(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Yellow(input...)))

	*c = ANSI(v)

	return c
}

// Blue applies the blue color to the input string(s) and returns it as an ANSI string.
// It uses the Blue() function from the ANSI package to convert the input to blue color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Blue(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Blue(input...)))

	*c = ANSI(v)

	return c
}

// Purple applies the purple color to the input string(s) and returns it as an ANSI string.
// It uses the Purple() function from the ANSI package to convert the input to purple color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Purple(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Purple(input...)))

	*c = ANSI(v)

	return c
}

// Cyan applies the cyan color to the input string(s) and returns it as an ANSI string.
// It uses the Cyan() function from the ANSI package to convert the input to cyan color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Cyan(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Cyan(input...)))

	*c = ANSI(v)

	return c
}

// Gray applies the gray color to the input string(s) and returns it as an ANSI string.
// It uses the Gray() function from the ANSI package to convert the input to gray color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Gray(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Gray(input...)))

	*c = ANSI(v)

	return c
}

// White applies the white color to the input string(s) and returns it as an ANSI string.
// It uses the White() function from the ANSI package to convert the input to white color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) White(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, White(input...)))

	*c = ANSI(v)

	return c
}

// Default applies the default color to the input string(s) and returns it as an ANSI string.
// It uses the Default() function from the ANSI package to convert the input to default color.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Default(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Default(input...)))

	*c = ANSI(v)

	return c
}

// Bold applies the bold style to the input string(s) and returns it as an ANSI string.
// It uses the Bold() function from the ANSI package to convert the input to bold style.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Bold(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Bold(input...)))

	*c = ANSI(v)

	return c
}

// Dim applies a dimmed style to the input string(s) and returns it as an ANSI string.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Dim(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Dim(input...)))

	*c = ANSI(v)

	return c
}

// Italic applies the italic style to the input string(s) and returns it as an ANSI string.
// It uses the Italic() function from the ANSI package to convert the input to italic style.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Italic(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Italic(input...)))

	*c = ANSI(v)

	return c
}

// Print writes the ANSI string to the output buffer without a newline character.
func (c *ANSI) Print() {
	var v string

	if c != nil {
		v = string(*c)
	}

	fmt.Fprintf(os.Stdout, "%s ", v)
}

// Write writes the ANSI string to the output buffer with a newline character.
func (c *ANSI) Write() {
	var v string

	if c != nil {
		v = string(*c)
	}

	fmt.Fprintf(os.Stdout, "%s\n", v)
}

// Overload allows passing ANSI Escape characters directly.
// It constructs an ANSI string with the provided ANSI Escape characters and the input string.
func (c *ANSI) Overload(ansi []byte, input string) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Overload(ansi, input)))

	*c = ANSI(v)

	return c
}

// String returns the ANSI string as a raw string.
func (c *ANSI) String() string {
	if c == nil {
		return ""
	}

	return string(*c)
}

// Color initializes and returns a new ANSI string.
func Color() *ANSI {
	return new(ANSI)
}

// typecast converts the provided entity to a string. Currently supports string and int types.
func typecast(entity interface{}) string {
	var partial string

	if cast, ok := entity.(string); ok {
		partial = fmt.Sprintf("%s", cast)
	} else if cast, ok := entity.(int); ok {
		partial = fmt.Sprintf("%d", cast)
	}

	return partial
}

// Overload constructs an ANSI string with the provided ANSI Escape characters and the input string.
func Overload(ansi []byte, input string) string {
	output := make([]string, 0)

	var partials []string

	color := ansi

	if Available() {
		partials = []string{string(color), input, reset}
	} else {
		partials = []string{input}
	}

	output = append(output, strings.Join(partials, ""))

	return strings.Join(output, " ")
}

// Default applies default color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Default(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := reset

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Black applies the black color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Black(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := black

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Red applies the red color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Red(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := red

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Green applies the green color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Green(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := green

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Yellow applies the yellow color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Yellow(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := yellow

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Blue applies the blue color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Blue(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := blue

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Purple applies the purple color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Purple(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := purple

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Cyan applies the cyan color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Cyan(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := cyan

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Gray applies the gray color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func Gray(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := gray

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// White applies the white color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the color code before and the reset code after each input string.
func White(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := white

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Bold applies the bold style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func Bold(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := bold

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Italic applies the italic style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func Italic(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := italic

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Dim applies a dimmed style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and ci is false, it adds the style code before and the reset code after each input string.
func Dim(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := dim

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Available checks if the terminal has a TTY (teletypewriter) available and returns a boolean value indicating
// if the system's output buffer is capable of color output.
func Available() bool {
	return !(ci()) && runtime.GOOS != "windows"
}

var force atomic.Value

// Force will force the runtime to color output.
func Force() {
	force.Store(true)
}

func init() {
	force.Store(false)
}

// // init initializes the ci variable based on the environment or terminal support for ANSI colors.
func ci() bool {
	// --> force color => ci = false
	force := force.Load().(bool)

	switch {
	case force:
		return false
	case os.Getenv("CI") == "true":
		return true
	case os.Getenv("CI") == "false":
		return false
	case term.IsTerminal(int(os.Stdout.Fd())):
		return false
	default:
		return true
	}
}
