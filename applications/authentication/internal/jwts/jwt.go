package jwts

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"http-api/internal/logging"
	secrets "http-api/internal/providers/kubernetes/secrets/ethr"
)

var (
	ethr   secrets.ETHR
	secret []byte
)

func init() {
	// if exception := godotenv.Load(); exception != nil {
	//	panic(exception)
	// }

	ethr = secrets.Secret()

	secret = []byte(ethr.Secret())
}

// Claims https://www.iana.org/assignments/jwt/jwt.xhtml
type Claims struct {
	jwt.RegisteredClaims

	// UID - User ID
	UID int `json:"uid"`
	// Email - User Email
	Email string `json:"email"`
	// Verified - User verification
	Verified bool `json:"verified"`
}

func (c *Claims) String() string {
	return string(c.Bytes())
}

func (c *Claims) Bytes() []byte {
	buffer, exception := json.MarshalIndent(c, "", "    ")
	if exception != nil {
		panic(exception)
	}

	return buffer
}

func (c *Claims) token() *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, c)
}

func (c *Claims) Sign() (signature string, exception error) {
	return c.token().SignedString(secret)
}

func Validate(ctx context.Context, jwtstring string) (claims *Claims, e error) {
	token, e := jwt.ParseWithClaims(jwtstring, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithLeeway(5*time.Second), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer(Issuer), jwt.WithAudience(Audience))

	if e != nil {
		slog.Log(ctx, slog.LevelWarn, "Error Parsing JWT Claims", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(e.Error()),
		})

		return nil, e
	}

	if authorization, ok := token.Claims.(*Claims); ok && token.Valid {
		// @todo create test case
		if authorization.Email != "" {
			return authorization, nil
		}

		slog.Log(ctx, slog.LevelWarn, "Invalid Email Address in Claims Data", logging.Structure("exception", authorization))

		return nil, errors.New("JWT Verification Failure")
	}

	return nil, errors.New("JWT Verification Failure")
}
