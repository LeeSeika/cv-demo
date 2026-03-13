package auth

import (
	"time"

	"github.com/pascaldekloe/jwt"
)

type Auth interface {
	// ParseToken gets jwt.Claims and auth info from token string
	ParseToken(token string) (*jwt.Claims, Info, error)

	// GenAccessToken generates token
	GenToken(info Info, opts ...GenTokenOpt) (string, error)

	// CheckToken verify if the token has expired and construct the auth info
	CheckToken(token string) (bool, Info, string, error)

	// AccessTokenExpiry returns access token expiry
	AccessTokenExpiry() time.Duration

	// VerificationExpiry returns verification token expiry
	VerificationTokenExpiry() time.Duration
}
