package oauth2

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/big"

	"github.com/leeseika/cv-demo/pkg/model/dto"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

const (
	saltLength = 16
)

func (o *oauth2) CreateAuthSession(ctx context.Context, authInfo jsonmodel.AuthInfo, payload dto.OAuthPayload) (string, string, error) {
	if o.db == nil {
		salt, err := generateRandomString(saltLength)
		if err != nil {
			return "", "", err
		}

		raw := salt + payload.ClientID + payload.CodeChallenge + payload.CodeChallengeMethod + payload.RedirectURI + payload.ResponseType + payload.Scope + payload.State
		hashed := generateHash(raw)

		return salt, hashed, nil
	}

	var shopApp object.ShopApp
	err := o.db.First(&shopApp, "app_id = ? AND shop_id = ?", payload.ClientID, authInfo.ShopID).Error
	if err != nil {
		return "", "", err
	}

	if shopApp.Status != "installed" {
		return "", "", errors.New("app not installed")
	}

	salt, err := generateRandomString(saltLength)
	if err != nil {
		return "", "", err
	}

	raw := salt + payload.ClientID + payload.CodeChallenge + payload.CodeChallengeMethod + payload.RedirectURI + payload.ResponseType + payload.Scope + payload.State
	hashed := generateHash(raw)

	return salt, hashed, nil
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	ret := ""
	max := big.NewInt(int64(len(letters)))

	for i := 0; i < n; i++ {
		num, err := crand.Int(crand.Reader, max)
		if err != nil {
			return "", err
		}
		ret = ret + string(letters[num.Int64()])
	}

	return ret, nil
}

func generateHash(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hash
}
