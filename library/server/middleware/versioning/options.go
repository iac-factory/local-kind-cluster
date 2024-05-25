package versioning

import "ethr.gg/server/internal/keystore"

type Settings struct {
	Version
}

type Variadic keystore.Variadic[Settings]
