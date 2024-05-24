package middleware

type Key string

func (k Key) String() string {
	return string(k)
}

type Store interface {
	Path() Key
	Service() Key
	Version() Key
	Telemetry() Key
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

var s = store{}

func Keys() Store {
	return s
}
