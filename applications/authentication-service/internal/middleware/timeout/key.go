package timeout

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Timeout()
