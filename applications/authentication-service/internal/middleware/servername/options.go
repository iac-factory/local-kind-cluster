package servername

import "authentication-service/internal/middleware/types"

type Settings struct {

	// Server represents the "Server" [http.Header].
	Server string `json:"server" yaml:"server"`
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
