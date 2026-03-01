package auth

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func RequirePageAuth(c *gin.Context) {
	tokenString, err := c.Cookie(authTokenCookieName)
	if err != nil || !isLoginTokenValid(tokenString) {
		next := c.Request.URL.Path
		if len(c.Request.URL.RawQuery) > 0 {
			next = next + "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusFound, "/auth/login?next="+url.QueryEscape(next))
		c.Abort()
		return
	}

	c.Next()
}

func RequireAPIAuth(c *gin.Context) {
	tokenString, err := c.Cookie(authTokenCookieName)
	if err != nil || !isLoginTokenValid(tokenString) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	c.Next()
}

func RequireOAuthAuth(c *gin.Context) {
	tokenString, err := c.Cookie(authTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	authInfo, err := parseLoginToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	if clientID := c.Query("client_id"); len(clientID) > 0 {
		authInfo.AppID = clientID
	}
	if clientID := c.PostForm("client_id"); len(clientID) > 0 {
		authInfo.AppID = clientID
	}

	if len(authInfo.ShopID) == 0 {
		authInfo.ShopID = "shop_demo"
	}
	if len(authInfo.Role) == 0 {
		authInfo.Role = "owner"
	}
	if len(authInfo.SessionID) == 0 {
		authInfo.SessionID = "sess_demo"
	}
	if len(authInfo.AccountID) == 0 {
		authInfo.AccountID = "acc_demo"
	}

	c.Set("auth_info", jsonmodel.AuthInfo(authInfo))
	c.Next()
}
