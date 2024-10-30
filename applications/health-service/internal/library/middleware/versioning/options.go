package versioning

import (
	"health-service/internal/library/middleware/types"
)

type Settings struct {
	// The `Version` struct represents the version information of a service or API. It has two fields: `API` and `Service`.
	Version Version `json:"version" yaml:"version"`
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
