package cors

import (
	"authentication-service/internal/middleware/types"
)

type Settings struct {
	Debug bool
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{
		Debug: false,
	}
}
