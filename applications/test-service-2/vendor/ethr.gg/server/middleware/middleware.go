package middleware

import (
	"ethr.gg/server/middleware/name"
	"ethr.gg/server/middleware/path"
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

func (g generic) Service() name.Implementation {
	return name.New()
}

func (generic) Telemetry() telemetry.Implementation {
	return telemetry.New()
}

var implementation = generic{}

type Middleware interface {
	Path() path.Implementation
	Version() versioning.Implementation
	Service() name.Implementation
	Telemetry() telemetry.Implementation
}

func New() Middleware {
	return implementation
}
