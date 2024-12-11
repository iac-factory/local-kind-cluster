package authentication

import (
	"net/http"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/authentication"

	"verification-service/internal/token"
)

func Middleware(next http.Handler) http.Handler {
	fn := middleware.New().Authentication().Configuration(func(options *authentication.Settings) {
		options.Verification = token.Verify
	})

	return fn.Middleware(next)
}
