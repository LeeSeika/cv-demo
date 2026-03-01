package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if len(secret) == 0 {
		secret = "auth-server-demo-secret"
	}
	return secret
}

func RequireAuth(c *gin.Context) {
	tokenString, err := tokenFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	authInfo, err := parseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	c.Set("auth_info", authInfo)
	c.Next()
}

func tokenFromRequest(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimPrefix(header, "Bearer ")
		if len(token) > 0 {
			return token, nil
		}
	}

	token, err := c.Cookie("auth_token")
	if err == nil && len(token) > 0 {
		return token, nil
	}

	return "", errors.New("missing token")
}

func parseToken(tokenString string) (jsonmodel.AuthInfo, error) {
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
