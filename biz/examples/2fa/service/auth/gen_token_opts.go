package auth

import (
	"time"

	"github.com/leeseika/cv-demo/pkg/constants"
	"github.com/pascaldekloe/jwt"
)

// GenTokenOpt gen access token option.
type GenTokenOpt func(c *jwt.Claims)

// GenWithSubject gen access token with subject.
func GenWithSubject(s constants.JWTSubject) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.Subject = string(s)
	}
}

// GenWithAudiences gen access token with audiences.
func GenWithAudiences(a []string) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.Audiences = a
	}
}

// GenWithExpires gen access token with expires.
func GenWithExpires(t time.Time) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.Expires = jwt.NewNumericTime(t)
	}
}

// GenWithNotBefore gen access token with not before.
func GenWithNotBefore(t time.Time) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.NotBefore = jwt.NewNumericTime(t)
	}
}

// GenWithIssued gen access token with issued.
func GenWithIssued(t time.Time) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.Issued = jwt.NewNumericTime(t)
	}
}

// GenWithID gen access token with id.
func GenWithID(id string) GenTokenOpt {
	return func(c *jwt.Claims) {
		c.ID = id
	}
}
