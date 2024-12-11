package versioning

import (
	"verification-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Version()
