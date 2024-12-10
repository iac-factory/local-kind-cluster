package token

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var signer []byte

func init() {
	value := os.Getenv("JWT_SIGNING_TOKEN")
	if value == "" {
		slog.Warn("No JWT_SIGNING_TOKEN Environment Variable Set... Defaulting to Development Token")

		value = "cnZ7Pc-xg20iP2qecFYj2bEt1O1qBDCfOmkdz5i6Fxw"
	}

	if len(value) == 0 {
		log.Fatal("invalid jwt signing token")
	}

	signer = []byte(value)
}

func Verify(ctx context.Context, t string) (*jwt.Token, error) {
	token, e := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		v, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		_ = v

		return signer, nil
	})

	if e != nil {
		slog.WarnContext(ctx, "Error Parsing JWT Token", slog.String("error", e.Error()), slog.String("jwt", t))
		return nil, e
	}

	switch {
	case token.Valid:
		slog.DebugContext(ctx, "Verified Valid Token")
		return token, nil
	case errors.Is(e, jwt.ErrTokenMalformed):
		slog.WarnContext(ctx, "Unable to Verify Malformed String as JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		slog.WarnContext(ctx, "Invalid JWT Signature", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenExpired):
		slog.WarnContext(ctx, "Expired JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenNotValidYet):
		slog.WarnContext(ctx, "Received a Future, Valid JWT Token", slog.String("error", e.Error()))
	default:
		slog.ErrorContext(ctx, "Unknown Error While Attempting to Validate JWT Token", slog.String("error", e.Error()))
	}

	return nil, e
}
