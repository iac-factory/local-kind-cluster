package name

import "ethr.gg/server/internal/middleware"

type Settings struct {
	Service string
}

type Variadic middleware.Variadic[Settings]
