package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leeseika/cv-demo/biz/design/oauth2/pkg/auth"
)

const bearerTokenLen = 2

func Authentication(c *gin.Context) {
	bToken := strings.Split(c.GetHeader("Authorization"), "Bearer ")
	if len(bToken) != bearerTokenLen {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token := bToken[1]

	claims, authInfo, err := auth.ParseToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("auth_info", authInfo)
	c.Set("claims", claims)

	c.Next()
}
