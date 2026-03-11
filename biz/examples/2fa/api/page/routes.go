package page

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter) {
	r.GET("/register", RegisterPage)
	r.GET("/register/qrcode", RegisterQRCodePage)

	r.GET("/login", LoginPage)

	r.GET("/otp", OTPPage)

	r.GET("/account", AccountPage)
}
