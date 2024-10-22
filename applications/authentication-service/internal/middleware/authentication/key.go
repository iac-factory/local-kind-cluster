package authentication

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Authentication()
