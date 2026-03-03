package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	imagesvc "github.com/leeseika/cv-demo/biz/design/image/storage/service"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func Preupload(c *gin.Context) {
	var req dto.ImagePreuploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	v, exists := c.Get("auth_info")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authInfo := v.(*jsonmodel.AuthInfo)
	shopID := authInfo.ShopID

	uploadURL, err := imagesvc.Get().Preupload(c, shopID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadURL,
		"method":     http.MethodPut,
	})
}
