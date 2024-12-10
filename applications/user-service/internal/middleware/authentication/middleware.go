package authentication

import (
	"net/http"

	"user-service/internal/library/middleware"
	"user-service/internal/library/middleware/authentication"

	"user-service/internal/token"
)

func Middleware(next http.Handler) http.Handler {
	fn := middleware.New().Authentication().Configuration(func(options *authentication.Settings) {
		options.Verification = token.Verify
	})

	return fn.Middleware(next)
}
