package server

import (
	"net/http"
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /health", Health)

	return mux
}
