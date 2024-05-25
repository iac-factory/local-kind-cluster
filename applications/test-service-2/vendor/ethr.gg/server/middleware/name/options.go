package name

import "ethr.gg/server/internal/keystore"

type Settings struct {
	Service string
}

type Variadic keystore.Variadic[Settings]
