package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

const authTokenCookieName = "auth_token"

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if len(secret) == 0 {
		secret = "auth-server-demo-secret"
	}
	return secret
}

func issueLoginToken(info jsonmodel.AuthInfo) (string, error) {
	claims := jwt.MapClaims{
		"account_id": info.AccountID,
		"shop_id":    info.ShopID,
		"session_id": info.SessionID,
		"role":       info.Role,
		"app_id":     info.AppID,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func parseLoginToken(tokenString string) (jsonmodel.AuthInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return jsonmodel.AuthInfo{}, err
	}
	if !token.Valid {
		return jsonmodel.AuthInfo{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jsonmodel.AuthInfo{}, errors.New("invalid claims")
	}

	getString := func(key string) string {
		if v, ok := claims[key].(string); ok {
			return v
		}
		return ""
	}

	return jsonmodel.AuthInfo{
		AccountID: getString("account_id"),
		ShopID:    getString("shop_id"),
		SessionID: getString("session_id"),
		Role:      getString("role"),
		AppID:     getString("app_id"),
	}, nil
}

func isLoginTokenValid(tokenString string) bool {
	_, err := parseLoginToken(tokenString)
	return err == nil
}
