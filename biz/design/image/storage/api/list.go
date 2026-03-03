package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	imagesvc "github.com/leeseika/cv-demo/biz/design/image/storage/service"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func List(c *gin.Context) {
	shopID := c.Query("shop_id")
	if len(shopID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop_id is required"})
		return
	}

	images, err := imagesvc.Get().List(c, shopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	urlBuilder := driver.GetStorageURLBuilder()
	imageResps := make([]*dto.ImageResponse, 0, len(images))
	for _, img := range images {
		imageResps = append(imageResps, dto.BuildImageResponse(img, urlBuilder.BuildURL(img.Bucket, img.FileKey)))
	}

	c.JSON(http.StatusOK, gin.H{
		"images": imageResps,
	})
}
