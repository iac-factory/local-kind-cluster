package versioning

import (
	"user-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Version()
