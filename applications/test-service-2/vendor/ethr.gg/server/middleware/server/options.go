package server

import "ethr.gg/server/internal/keystore"

type Settings struct {
	Server string
}

type Variadic keystore.Variadic[Settings]
