package jwts

import "time"

const (
	Issuer   = "Polygun-Authorization-Server"
	Audience = "https://ethr.gg"
	// Expiration - constant [time.Duration] for JWT expiration
	Expiration = 18 * time.Hour
)
