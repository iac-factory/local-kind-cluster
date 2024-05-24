package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server initializes a http.Server with application-specific configuration.
func Server(ctx context.Context, handler http.Handler, port string) *http.Server {
	return &http.Server{
		Addr:                         fmt.Sprintf("0.0.0.0:%s", port),
		Handler:                      otelhttp.handler,
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
