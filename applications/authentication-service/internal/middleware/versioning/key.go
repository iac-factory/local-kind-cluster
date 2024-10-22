package versioning

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Version()
