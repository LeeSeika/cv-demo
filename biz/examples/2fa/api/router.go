package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	accountapi "github.com/leeseika/cv-demo/biz/examples/2fa/api/account"
	pageapi "github.com/leeseika/cv-demo/biz/examples/2fa/api/page"
)

func (s *Server) addRoutes(engine *gin.Engine) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/page/login")
	})

	apiGroup := engine.Group("/api")
	s.registerAuth(apiGroup)
	s.registerAccount(apiGroup)

	pageGroup := engine.Group("/page")
	pageapi.RegisterRoutes(pageGroup)
}

func (s *Server) registerAuth(r *gin.RouterGroup) {
	r.POST("/auth/register", accountapi.Register)
	r.POST("/auth/login", accountapi.Login)
	r.POST("/auth/2fa/verify", accountapi.VerifyOTP)
}

func (s *Server) registerAccount(r *gin.RouterGroup) {
	accountGroup := r.Group("/account")
	accountapi.RegisterRoutes(accountGroup)
}
