package path

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Path()
