/*
Copyright: 2024, Deep Codify Limited

Enterprise Merchant Console
*/

package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/pascaldekloe/jwt"
)

const (
	rawPublicKey               = "-----BEGIN PUBLIC KEY-----\nMCowBQYDK2VwAyEAX6r2TxjxRw2I4KK914dV3OsZuI6T0dnE3xM83hlFwXg=\n-----END PUBLIC KEY-----"
	rawPrivateKey              = "-----BEGIN PRIVATE KEY-----\nMC4CAQAwBQYDK2VwBCIEIBZBDUmgH11eLWHrK9jZpJajO93vKlKkSKb7EZWIevOK\n-----END PRIVATE KEY-----"
	jwtAccessTokenExpiry       = 15 * time.Minute
	jwtVerificationTokenExpiry = 5 * time.Minute
)

type Info struct {
	AccountID string `json:"account_id"`
}

var _auth Auth

type authService struct {
	publicKey                  ed25519.PublicKey  // JWT Public Key
	privateKey                 ed25519.PrivateKey // JWT Private Key
	jwtAccessTokenExpiry       time.Duration      // JWT Access Token Expiry
	jwtVerificationTokenExpiry time.Duration      // JWT Verification Token Expiry
}

func Get() Auth {
	return _auth
}

func Init() {
	privateKey, err := decodePrivateKey(rawPrivateKey)
	if err != nil {
		panic("failed to decode private key: " + err.Error())
	}

	var publicKey ed25519.PublicKey
	if len(rawPublicKey) > 0 {
		publicKey, err = decodePublicKey(rawPublicKey)
		if err != nil {
			panic("failed to decode public key: " + err.Error())
		}
	} else {
		derived, ok := privateKey.Public().(ed25519.PublicKey)
		if !ok {
			panic("failed to derive public key")
		}
		publicKey = derived
	}
	a := authService{
		publicKey:                  publicKey,
		privateKey:                 privateKey,
		jwtAccessTokenExpiry:       jwtAccessTokenExpiry,
		jwtVerificationTokenExpiry: jwtVerificationTokenExpiry,
	}

	_auth = &a
}

// ParseToken gets jwt.Claims and auth info from token string.
func (as *authService) ParseToken(token string) (*jwt.Claims, Info, error) {
	c, err := jwt.EdDSACheck([]byte(token), as.publicKey)

	if err != nil {
		return nil, Info{}, err
	}

	var info Info

	if err := json.Unmarshal(c.Raw, &info); err != nil {
		return nil, Info{}, err
	}

	return c, info, nil
}

// GenToken generates token
func (as *authService) GenToken(info Info, opts ...GenTokenOpt) (string, error) {
	var c jwt.Claims

	for _, opt := range opts {
		opt(&c)
	}

	m, err := structToMap(info, "json")

	if err != nil {
		return "", err
	}

	c.Set = m
	c.Issued = jwt.NewNumericTime(time.Now().Round(time.Second))

	token, err := c.EdDSASign(as.privateKey)

	if err != nil {
		return "", err
	}

	return string(token), nil
}

// CheckToken verify if the token has expired and construct the auth info
func (as *authService) CheckToken(token string) (bool, Info, string, error) {
	expired := false

	c, err := jwt.EdDSACheck([]byte(token), as.publicKey)

	if err != nil {
		return expired, Info{}, "", err
	}

	if !c.Valid(time.Now()) {
		expired = true
		return expired, Info{}, "", nil
	}

	var info Info

	if err := json.Unmarshal(c.Raw, &info); err != nil {
		return expired, Info{}, "", err
	}

	return expired, info, c.Subject, nil
}

func decodePublicKey(pemEncoded string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	key, err := x509.ParsePKIXPublicKey(x509Encoded)

	if err != nil {
		return nil, err
	}

	ed25519Key, ok := key.(ed25519.PublicKey)

	if !ok {
		return nil, errors.New("invalid public key")
	}

	return ed25519Key, nil
}

func decodePrivateKey(pemEncoded string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	key, err := x509.ParsePKCS8PrivateKey(x509Encoded)

	if err != nil {
		return nil, err
	}

	ed25519Key, ok := key.(ed25519.PrivateKey)

	if !ok {
		return nil, errors.New("invalid private key")
	}

	return ed25519Key, nil
}

func structToMap(in interface{}, tag string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			m[tagv] = v.Field(i).Interface()
		}
	}
	return m, nil
}
