package state

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().State()
