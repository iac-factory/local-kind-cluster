package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"user-service/internal/api/avatar"
	"user-service/internal/api/delete"
	"user-service/internal/api/me"
	"user-service/internal/api/registration"
	"user-service/internal/library/server"

	"user-service/internal/middleware/authentication"
)

func Router(parent *http.ServeMux) {
	{ // --> authentication endpoints
		parent.Handle("GET /@me", authentication.Middleware(otelhttp.WithRouteTag("/@me", me.Handler)))

		parent.Handle("DELETE /users/{id}", authentication.Middleware(otelhttp.WithRouteTag("/{id}", delete.Handler)))
		parent.Handle("PATCH /users/{id}/avatar", authentication.Middleware(otelhttp.WithRouteTag("/avatar", avatar.Handler)))
	}

	parent.HandleFunc("GET /health", server.Health)

	parent.Handle("POST /register", otelhttp.WithRouteTag("/register", registration.Handler))
}
