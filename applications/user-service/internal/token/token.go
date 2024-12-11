package token

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"

	"github.com/golang-jwt/jwt/v5"

	"user-service/internal/library/middleware"
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

// Claims is a standard [jwt.RegisteredClaims] structure that can be extended with additional, custom claims data.
type Claims struct {
	jwt.RegisteredClaims
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
		slog.DebugContext(ctx, "Basic Token Parsing was Successful - Vetting Additional Claims")

		// Verify the token's target audience(s).
		audiences, e := token.Claims.GetAudience()
		if e != nil {
			slog.ErrorContext(ctx, "Error Parsing Audience Claims", slog.String("error", e.Error()), slog.Any("claims", token.Claims))
			e = fmt.Errorf("unable to parse audience claims: %w", e)
			return nil, e
		}

		service := middleware.New().Service().Value(ctx)
		if !(slices.Contains(audiences, service)) {
			slog.WarnContext(ctx, "JWT Claims Don't Contain Applicable Audience - Invalidating", slog.Any("claims", token.Claims))
			e = jwt.ErrTokenInvalidAudience
			return nil, e
		}

		return token, nil
	case errors.Is(e, jwt.ErrTokenMalformed):
		slog.WarnContext(ctx, "Unable to Verify Malformed String as JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenSignatureInvalid):
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
