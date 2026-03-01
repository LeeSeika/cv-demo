package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ordersvc "github.com/leeseika/cv-demo/biz/design/oauth2/resource-server/service/order"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
)

func List(c *gin.Context) {
	v, ok := c.Get("auth_info")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authInfo, ok := v.(jsonmodel.AuthInfo)
	if !ok || len(authInfo.ShopID) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	orders, err := ordersvc.Get().ListByShopID(c, authInfo.ShopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shop_id": authInfo.ShopID,
		"orders":  orders,
	})
}
