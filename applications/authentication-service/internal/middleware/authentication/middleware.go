package authentication

import (
	"net/http"

	"authentication-service/internal/library/middleware"
	"authentication-service/internal/library/middleware/authentication"

	"authentication-service/internal/token"
)

func Middleware(next http.Handler) http.Handler {
	fn := middleware.New().Authentication().Configuration(func(options *authentication.Settings) {
		options.Verification = token.Verify
	})

	return fn.Middleware(next)
}
