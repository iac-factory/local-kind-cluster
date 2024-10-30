package cors

import (
	"authentication-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().CORS()
