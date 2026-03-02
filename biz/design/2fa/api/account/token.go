package account

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func bearerToken(c *gin.Context) string {
	authorization := c.GetHeader("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	}

	return ""
}
