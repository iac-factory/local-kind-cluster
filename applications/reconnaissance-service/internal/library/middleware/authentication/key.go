package authentication

import (
	"reconnaissance-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().Authentication()
