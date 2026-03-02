package constants

type JWTSubject string

const (
	// JWTSubjectTOTPVerification indicates the JWT subject in verification token
	JWTSubjectTOTPVerification JWTSubject = "acc-totp-verification"

	// JWTSubjectLogin indicates the JWT subject in access token
	JWTSubjectAccess JWTSubject = "acc-access"
)
