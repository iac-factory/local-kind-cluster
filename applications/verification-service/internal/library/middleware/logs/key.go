package logs

import (
	"verification-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Logs()
