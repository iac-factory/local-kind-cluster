package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"ethr.gg/server/middleware"
	"ethr.gg/server/middleware/name"
	"ethr.gg/server/middleware/server"
	"ethr.gg/server/middleware/versioning"
)

// Server initializes a http.Server with application-specific configuration.
func Server(ctx context.Context, handler http.Handler, service, version, port string) *http.Server {
	handler = middleware.New().Server().Configuration(func(options *server.Settings) { options.Server = "ETHR-API-Server" }).Middleware(handler)

	handler = middleware.New().Version().Configuration(func(options *versioning.Settings) {
		options.Version.API = os.Getenv("VERSION")
		if options.Version.API == "" && os.Getenv("CI") == "" {
			options.Version.API = "local"
		}

		options.Version.Service = version
	}).Middleware(handler)

	handler = middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = service }).Middleware(handler)

	handler = middleware.New().Telemetry().Middleware(handler) // needs to be evaluated after path
	handler = middleware.New().Path().Middleware(handler)      // needs to be evaluated first

	return &http.Server{
		Addr:                         fmt.Sprintf("0.0.0.0:%s", port),
		Handler:                      handler,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  15 * time.Second,
		WriteTimeout:                 60 * time.Second,
		IdleTimeout:                  30 * time.Second,
		MaxHeaderBytes:               http.DefaultMaxHeaderBytes,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ConnContext: nil,
	}
}
