package account

import (
	"context"
	"errors"
	"time"

	"github.com/jltorresm/otpgo"
	"github.com/leeseika/cv-demo/biz/design/2fa/service/auth"
	"github.com/leeseika/cv-demo/pkg/constants"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/rs/zerolog/log"
)

func (a *account) VerifyOTP(ctx context.Context, accountID string, otp string) (string, error) {
	var account object.Account
	err := a.db.WithContext(ctx).Where("id = ?", accountID).First(&account).Error
	if err != nil {
		return "", errors.New("account not found")
	}

	// verify OTP
	key := account.TOTPSecret
	totp := otpgo.TOTP{
		Key: key,
	}

	verified, err := totp.Validate(otp)
	if err != nil || !verified {
		log.Err(err).Msg("failed to verify otp")
		return "", errors.New("invalid verification code")
	}

	authInfo := auth.Info{
		AccountID: account.ID,
	}
	token, err := auth.Get().GenToken(
		authInfo,
		auth.GenWithExpires(time.Now().Add(auth.Get().AccessTokenExpiry())),
		auth.GenWithSubject(constants.JWTSubjectAccess),
	)
	if err != nil {
		log.Err(err).Msg("failed to generate access token")
		return "", errors.New("failed to generate access token")
	}

	return token, nil
}
