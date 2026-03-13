package auth

import "time"

// AccessTokenExpiry returns access token expiry
func (as *authService) AccessTokenExpiry() time.Duration {
	return as.jwtAccessTokenExpiry
}

// VerificationExpiry returns verification token expiry
func (as *authService) VerificationTokenExpiry() time.Duration {
	return as.jwtVerificationTokenExpiry
}
