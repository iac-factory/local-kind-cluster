package timeout

import (
	"health-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Timeout()
