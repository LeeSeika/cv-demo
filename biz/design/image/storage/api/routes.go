package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter) {
	r.POST("/images", Preupload)
	r.POST("/images/:id/confirm", ConfirmUpload)
	r.GET("/images", List)
}
