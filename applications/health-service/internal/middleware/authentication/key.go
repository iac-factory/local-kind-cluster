package authentication

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().Authentication()
