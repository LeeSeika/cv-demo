package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appsvc "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/app"
)

func List(c *gin.Context) {
	apps, err := appsvc.Get().List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list apps"})
		return
	}

	c.JSON(http.StatusOK, apps)
}
