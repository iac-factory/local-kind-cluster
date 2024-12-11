package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"verification-service/internal/api/deletion"
	"verification-service/internal/api/register"
	"verification-service/internal/api/status"
	"verification-service/internal/api/verify"
	"verification-service/internal/library/middleware"
	"verification-service/internal/library/server"
)

func Router(parent *http.ServeMux) {
	var authentication = func(parent *http.ServeMux) {
		middlewares := middleware.Middleware()
		middlewares.Add(middleware.New().Authentication().Middleware)

		mux := http.NewServeMux()

		mux.Handle("DELETE /", otelhttp.WithRouteTag("/", deletion.Handler))
		mux.Handle("POST /register", otelhttp.WithRouteTag("/register", register.Handler))
		mux.Handle("POST /verify", otelhttp.WithRouteTag("/verify", verify.Handler))
		mux.Handle("GET /status", otelhttp.WithRouteTag("/status", status.Handler))

		handler := middlewares.Handler(mux)

		parent.Handle("/", handler)
	}

	authentication(parent)

	parent.HandleFunc("GET /health", server.Health)
}
