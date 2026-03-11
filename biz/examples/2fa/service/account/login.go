package account

import (
	"context"
	"errors"
	"time"

	"github.com/leeseika/cv-demo/biz/examples/2fa/service/auth"
	"github.com/leeseika/cv-demo/pkg/constants"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) Login(ctx context.Context, email, password string) (string, error) {
	var account object.Account
	err := a.db.WithContext(ctx).Where("email = ? AND password = ?", email, password).First(&account).Error
	if err != nil {
		return "", errors.New("email or password incorrect")
	}

	authInfo := auth.Info{
		AccountID: account.ID,
	}

	// generate 2fa verification token
	token, err := auth.Get().GenToken(
		authInfo,
		auth.GenWithExpires(time.Now().Add(auth.Get().VerificationTokenExpiry())),
		auth.GenWithSubject(constants.JWTSubjectTOTPVerification),
	)

	if err != nil {
		return "", errors.New("failed to generate verification token")
	}

	return token, nil
}
