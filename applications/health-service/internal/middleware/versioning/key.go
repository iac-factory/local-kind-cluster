package versioning

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().Version()
