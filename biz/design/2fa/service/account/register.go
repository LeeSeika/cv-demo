package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jltorresm/otpgo"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) Register(ctx context.Context, email, password string) (string, error) {
	// Check if email already exists
	var count int64
	err := a.db.WithContext(ctx).Model(&object.Account{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "", fmt.Errorf("email %s already registered", email)
	}

	totp := otpgo.TOTP{}
	_, err = totp.Generate()
	if err != nil {
		return "", err
	}
	base64EncodedQRImage, err := totp.KeyUri(email, "http://localhost:9100").QRCode()
	if err != nil {
		return "", err
	}

	id := uuid.NewString()
	account := object.Account{
		ID:         id,
		Email:      email,
		Password:   password,
		TOTPSecret: totp.Key,
	}

	err = a.db.WithContext(ctx).Create(&account).Error
	if err != nil {
		return "", err
	}

	return base64EncodedQRImage, nil
}
