package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"authentication-service/internal/library/middleware"

	"authentication-service/internal/api/deletion"
	"authentication-service/internal/api/login"
	"authentication-service/internal/api/logout"
	"authentication-service/internal/api/refresh"
	"authentication-service/internal/api/registration"
	"authentication-service/internal/api/remove"
	"authentication-service/internal/api/session"
	"authentication-service/internal/api/status"
)

func Router(parent *http.ServeMux) {
	var authentication = func(parent *http.ServeMux) {
		middlewares := middleware.Middleware()
		middlewares.Add(middleware.New().Authentication().Middleware)

		mux := http.NewServeMux()

		mux.Handle("POST /refresh", otelhttp.WithRouteTag("/refresh", refresh.Handler))
		mux.Handle("GET /session", otelhttp.WithRouteTag("/session", session.Handler))

		mux.Handle("DELETE /{id}", otelhttp.WithRouteTag("/{id}", remove.Handler))
		mux.Handle("DELETE /{id}/hard", otelhttp.WithRouteTag("/{id}/hard", deletion.Handler))

		handler := middlewares.Handler(mux)

		parent.Handle("/", handler)
	}

	authentication(parent)

	parent.Handle("POST /login", otelhttp.WithRouteTag("/login", login.Handler))

	parent.Handle("GET /logout", otelhttp.WithRouteTag("/logout", logout.Handler))

	parent.Handle("POST /register", otelhttp.WithRouteTag("/register", registration.Handler))

	parent.Handle("GET /status", otelhttp.WithRouteTag("/status", status.Handler))
}
