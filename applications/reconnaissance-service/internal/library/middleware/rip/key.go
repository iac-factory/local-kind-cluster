package rip

import (
	"reconnaissance-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().RIP()
