package telemetrics

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Telemetry()
