package app

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	appsvc "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/app"
	"gorm.io/gorm"
)

func GetByID(c *gin.Context) {
	appID := c.Param("id")
	if len(appID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
		return
	}

	app, err := appsvc.Get().GetByID(c, appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get app"})
		return
	}

	c.JSON(http.StatusOK, app)
}
