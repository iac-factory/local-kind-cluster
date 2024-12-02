package path

import (
	"reconnaissance-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Path()
