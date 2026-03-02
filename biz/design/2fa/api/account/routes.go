package account

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter) {
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/verify-otp", VerifyOTP)
	r.GET("/account-info", GetAccountInfo)
}
