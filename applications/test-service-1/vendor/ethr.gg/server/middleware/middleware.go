package middleware

import (
	"ethr.gg/server/middleware/name"
	"ethr.gg/server/middleware/path"
	"ethr.gg/server/middleware/server"
	"ethr.gg/server/middleware/telemetry"
	"ethr.gg/server/middleware/versioning"
)

type generic struct{}

func (generic) Path() path.Implementation {
	return path.New()
}

func (generic) Version() versioning.Implementation {
	return versioning.New()
}

func (generic) Service() name.Implementation {
	return name.New()
}

func (generic) Telemetry() telemetry.Implementation {
	return telemetry.New()
}

func (generic) Server() server.Implementation {
	return server.New()
}

var implementation = generic{}

type Middleware interface {
	Path() path.Implementation           // Path - See the [path] package for additional details.
	Version() versioning.Implementation  // Version - See the [versioning] package for additional details.
	Service() name.Implementation        // Service - See the [name] package for additional details.
	Telemetry() telemetry.Implementation // Telemetry - See the [telemetry] package for additional details.
	Server() server.Implementation       // Server - See the [server] package for additional details.
}

func New() Middleware {
	return implementation
}
