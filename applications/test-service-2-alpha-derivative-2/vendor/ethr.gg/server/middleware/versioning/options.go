package versioning

import "ethr.gg/server/internal/middleware"

type Settings struct {
	Version string
}

type Variadic middleware.Variadic[Settings]
