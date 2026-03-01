package oauth2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-session/session/v3"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/oauth2"
	"github.com/leeseika/cv-demo/biz/design/oauth2/pkg/constants"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func Authorize(c *gin.Context) {
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

	authorizeReq, err := oauth2.Get().OAuth2Server().ValidationAuthorizeRequest(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	csrfToken := c.Request.FormValue("csrf_token")

	store, err := session.Start(c, c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session"})
		return
	}

	sessionShopID, ok := store.Get(constants.AppAuthSessionNameKey)
	if !ok || sessionShopID != authInfo.ShopID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionHash, ok := store.Get(constants.AppAuthSessionHashKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessionHashStr, ok := sessionHash.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = oauth2.Get().ValidateCSRFToken(c, csrfToken, sessionHashStr, authorizeReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = oauth2.Get().OAuth2Server().HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle authorize request"})
		return
	}
}
