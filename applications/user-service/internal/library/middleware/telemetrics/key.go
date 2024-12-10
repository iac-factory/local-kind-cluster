package telemetrics

import (
	"user-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Telemetry()
