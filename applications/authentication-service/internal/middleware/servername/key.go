package servername

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Server()
