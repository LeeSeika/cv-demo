package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter) {
	r.POST("/images", Preupload)
	r.GET("/images", List)
}
