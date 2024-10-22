package servername

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().Server()
