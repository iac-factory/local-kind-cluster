package tracing

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Tracer()
