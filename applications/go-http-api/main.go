package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.DebugContext(ctx, "Received Request")

	// Create an instance of the response structure
	response := Response{
		Message: "Hello, this is a JSON response",
		Status:  200,
	}

	// Set the response header content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Marshal the structure into JSON and write it to the response
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Handle requests at the root path using the handler function
	http.HandleFunc("GET /", handler)

	http.ListenAndServe(":5000", nil)
}
