package server

import (
	"fmt"
	"net/http"
	"sync"
)

// Internal represents an optional, private error structure. Internal is useful for additional logging.
type Internal struct {
	Error   error
	Message string
}

// Exception represents an error that includes an HTTP status, a code, and a descriptive message.
type Exception struct {
	Code    int    // Code represents the numeric HTTP status code associated with the Exception. See the [http] package for various status codes.
	Status  string // Status represents the HTTP status text associated with the Exception. See [http.StatusText] for value assignment. If Status isn't provided, it's automatically calculated using [http.StatusText] on the Code attribute.
	Message string // An optional, custom error message. This value will replace the error message used in [http.Error] if not empty, and means will be publicly exposed to end-users.

	Internal *Internal // Internal represents an optional, private error structure. Internal is useful for additional logging.
}

// Hook represents a function type that accepts a variadic number of functions as arguments, primarily used for executing hooks or callbacks.
type Hook func(args ...func())

// validate ensures the Exception struct has a valid Status based on its Code. Defaults to 500 Internal Server Error if the Code is invalid.
func (e *Exception) validate() {
	if e.Status == "" {
		e.Status = http.StatusText(e.Code)
	}

	if e.Status == "" {
		original := e.Code

		e.Code = http.StatusInternalServerError
		e.Status = http.StatusText(http.StatusInternalServerError)
		e.Message = fmt.Sprintf("Invalid HTTP Status-Code Provided: (%d) - Default to Internal-Server-Error Properties", original)
	}
}

// Error returns a formatted string representation of the Exception, including its code, status, and message.
func (e *Exception) Error() string {
	e.validate()

	if e.Message != "" {
		return fmt.Sprintf("%d %s: %s", e.Code, e.Status, e.Message)
	}

	return fmt.Sprintf("%s", e.Status)
}

// Response sends an HTTP response based on the Exception's properties and code, optionally executing hooks concurrently before returning the response.
func (e *Exception) Response(w http.ResponseWriter, hooks ...Hook) {
	e.validate()

	var wg sync.WaitGroup

	// --> Add the number of hooks to the WaitGroup.
	wg.Add(len(hooks))

	for index := range hooks {
		hook := hooks[index]

		// --> Run each task in a separate goroutine.
		go func(h Hook) {
			defer wg.Done() // Mark as done when the task is complete

			h()
		}(hook)
	}

	// --> Wait for all tasks to complete
	wg.Wait()

	if e.Message != "" {
		http.Error(w, e.Message, e.Code)

		return
	}

	http.Error(w, e.Status, e.Code)
}
