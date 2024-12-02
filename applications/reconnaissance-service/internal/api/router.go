package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"reconnaissance-service/internal/api/tls/expiration"
)

func Router(parent *http.ServeMux) {
	// var authentication = func(parent *http.ServeMux) {
	// 	middlewares := middleware.Middleware()
	// 	middlewares.Add(middleware.New().Authentication().Middleware)
	//
	// 	mux := http.NewServeMux()
	//
	// 	mux.Handle("POST /refresh", otelhttp.WithRouteTag("/refresh", refresh.Handler))
	// 	mux.Handle("GET /session", otelhttp.WithRouteTag("/session", session.Handler))
	//
	// 	mux.Handle("DELETE /{id}", otelhttp.WithRouteTag("/{id}", remove.Handler))
	// 	mux.Handle("DELETE /{id}/hard", otelhttp.WithRouteTag("/{id}/hard", deletion.Handler))
	//
	// 	handler := middlewares.Handler(mux)
	//
	// 	parent.Handle("/", handler)
	// }
	//
	// authentication(parent)

	parent.Handle("POST /tls/expiration", otelhttp.WithRouteTag("/tls/expiration", expiration.Handler))
}
