package account

import (
	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/legacy/service/account"
	"github.com/leeseika/cv-demo/pkg/utils/api"
)

func GetByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	data := gin.H{
		"id":         info.ID,
		"name":       info.Name,
		"avatar_url": info.AvatarURL,
	}
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", data))
}
