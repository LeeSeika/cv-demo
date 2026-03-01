package oauth2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/oauth2"
)

func Token(c *gin.Context) {
	err := oauth2.Get().OAuth2Server().HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle token request"})
		return
	}
}
