package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	imagesvc "github.com/leeseika/cv-demo/biz/design/image/storage/service"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func ConfirmUpload(c *gin.Context) {
	imageID := strings.TrimSpace(c.Param("id"))
	if len(imageID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image id is required"})
		return
	}

	v, exists := c.Get("auth_info")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authInfo := v.(*jsonmodel.AuthInfo)

	if err := imagesvc.Get().ConfirmUpload(c, authInfo.ShopID, imageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}