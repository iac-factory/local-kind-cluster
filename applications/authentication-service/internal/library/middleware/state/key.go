package state

import (
	"authentication-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().State()
