package rip

import (
	"health-service/internal/library/middleware/keystore"
)

var key = keystore.Keys().RIP()
