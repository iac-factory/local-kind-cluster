package keystore

// Key represents a constant context-key string value.
type Key string

func (k Key) String() string {
	return string(k)
}

// Store represents the interface that providers all package-specific context, context keys.
type Store interface {
	// Path represents the context.Context key: "path". See path.Implementation for the middleware.
	Path() Key

	// Service represents the context.Context key: "service". See name.Implementation for the middleware.
	Service() Key

	// Version represents the context.Context key: "version". See versioning.Implementation for the middleware.
	//
	//   - Used for configuring middleware that adds versioning information to both context keys and response headers.
	Version() Key

	// Telemetry represents the context.Context key: "telemetry". See telemetry.Implementation for the middleware.
	//
	//   - Used for configuring middleware that adds route-specific telemetry.
	Telemetry() Key

	// Server represents the context.Context key: "server". See server.Implementation for the middleware.
	//
	//   - Used for configuring middleware that sets the "Server" response header.
	Server() Key
}

type store struct{}

func (s store) Path() Key {
	return "path"
}

func (s store) Service() Key {
	return "service"
}

func (s store) Version() Key {
	return "version"
}

func (s store) Telemetry() Key {
	return "telemetry"
}

func (s store) Server() Key {
	return "server"
}

var s = store{}

func Keys() Store {
	return s
}
