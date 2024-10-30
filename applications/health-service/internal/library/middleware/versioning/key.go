package versioning

import (
	"health-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Version()
