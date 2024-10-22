package envoy

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().Envoy()
