package logs

import "authentication-service/internal/middleware/keystore"

var key = keystore.Keys().Logs()
