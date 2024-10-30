package tracing

import (
	"health-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Tracer()
