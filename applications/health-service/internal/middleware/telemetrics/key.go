package telemetrics

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().Telemetry()
