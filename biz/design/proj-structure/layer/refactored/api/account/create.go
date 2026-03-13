package account

import (
	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/refactored/service/account"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/utils/api"
)

func CreateAccount(c *gin.Context) {
	var req dto.CreateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, api.BadRequestResponse("Invalid request"))
		return
	}
	id, err := accountsvc.Get().CreateAccount(c, &req)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	c.JSON(200, api.SuccessResponse("Account created successfully", gin.H{"id": id}))
}
