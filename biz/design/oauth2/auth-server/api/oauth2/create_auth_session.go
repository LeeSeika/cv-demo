package oauth2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	session "github.com/go-session/session/v3"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/oauth2"
	"github.com/leeseika/cv-demo/biz/design/oauth2/pkg/constants"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func CreateAuthSession(c *gin.Context) {
	v, ok := c.Get("auth_info")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authInfo, ok := v.(jsonmodel.AuthInfo)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload dto.OAuthPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	csrfToken, hashed, err := oauth2.Get().CreateAuthSession(c, authInfo, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create auth session"})
		return
	}

	// set session
	store, err := session.Start(c, c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	store.Set(constants.AppAuthSessionNameKey, authInfo.ShopID)
	store.Set(constants.AppAuthSessionHashKey, hashed)
	err = store.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"csrf_token": csrfToken})
}
