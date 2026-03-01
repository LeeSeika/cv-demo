package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"time"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/pascaldekloe/jwt"
)

const (
	pemEncoded = "-----BEGIN PUBLIC KEY-----\nMCowBQYDK2VwAyEAX6r2TxjxRw2I4KK914dV3OsZuI6T0dnE3xM83hlFwXg=\n-----END PUBLIC KEY-----"
)

var (
	publicKey ed25519.PublicKey
)

func init() {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	key, err := x509.ParsePKIXPublicKey(x509Encoded)

	if err != nil {
		panic("failed to parse public key: " + err.Error())
	}

	ed25519Key, ok := key.(ed25519.PublicKey)

	if !ok {
		panic("invalid public key")
	}

	publicKey = ed25519Key
}

func ParseToken(token string) (*jwt.Claims, jsonmodel.AuthInfo, error) {
	c, err := jwt.EdDSACheck([]byte(token), publicKey)

	if err != nil {
		return nil, jsonmodel.AuthInfo{}, err
	}

	if !c.Valid(time.Now()) {
		return nil, jsonmodel.AuthInfo{}, errors.New("token expired")
	}

	var info jsonmodel.AuthInfo

	if err := json.Unmarshal(c.Raw, &info); err != nil {
		return nil, jsonmodel.AuthInfo{}, err
	}

	return c, info, nil
}
