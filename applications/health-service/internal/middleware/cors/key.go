package cors

import "health-service/internal/middleware/keystore"

var key = keystore.Keys().CORS()
