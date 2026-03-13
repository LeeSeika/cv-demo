package account

import (
	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/refactored/service/account"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/utils/api"
)

func Update(c *gin.Context) {
	var req dto.UpdateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, api.BadRequestResponse("Invalid request"))
		return
	}
	id := c.Param("id")
	err := accountsvc.Get().UpdateAccount(c, id, &req)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	c.JSON(200, api.SuccessResponse("Account updated successfully", nil))
}
