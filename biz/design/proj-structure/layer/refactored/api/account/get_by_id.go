package account

import (
	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/refactored/service/account"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/utils/api"
)

func GetByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountInfoByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// 提取需要返回的字段
	responseAccountInfo := &dto.AccountInfo{
		ID:        info.ID,
		Name:      info.Name,
		AvatarURL: info.AvatarURL,
	}
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", responseAccountInfo))
}
