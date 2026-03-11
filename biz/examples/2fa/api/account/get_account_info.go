package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/examples/2fa/service/account"
	authsvc "github.com/leeseika/cv-demo/biz/examples/2fa/service/auth"
	"github.com/leeseika/cv-demo/pkg/constants"
)

func GetAccountInfo(c *gin.Context) {
	token := bearerToken(c)
	if len(token) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token required"})
		return
	}

	expired, info, subject, err := authsvc.Get().CheckToken(token)
	if err != nil || expired || subject != string(constants.JWTSubjectAccess) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}
	if len(info.AccountID) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}

	account, err := accountsvc.Get().GetAccountInfo(c, info.AccountID)
	if err != nil {
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         account.ID,
		"email":      account.Email,
		"created_at": account.CreatedAt,
		"password":   account.Password,
		"updated_at": account.UpdatedAt,
	})
}
