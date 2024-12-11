package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"authentication-service/internal/api/delete"
	"authentication-service/internal/api/login"
	"authentication-service/internal/api/logout"
	"authentication-service/internal/api/refresh"
	"authentication-service/internal/api/registration"
	"authentication-service/internal/api/session"
	"authentication-service/internal/middleware/authentication"
)

func Router(parent *http.ServeMux) {
	{ // --> authentication endpoints
		parent.Handle("POST /refresh", authentication.Middleware(otelhttp.WithRouteTag("/refresh", refresh.Handler)))
		parent.Handle("GET /session", authentication.Middleware(otelhttp.WithRouteTag("/session", session.Handler)))
		parent.Handle("DELETE /users/{id}", authentication.Middleware(otelhttp.WithRouteTag("/", delete.Handler)))
	}

	parent.Handle("POST /login", otelhttp.WithRouteTag("/login", login.Handler))

	parent.Handle("GET /logout", otelhttp.WithRouteTag("/logout", logout.Handler))

	parent.Handle("POST /register", otelhttp.WithRouteTag("/register", registration.Handler))
}
