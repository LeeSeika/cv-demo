package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	appapi "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/api/app"
	authapi "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/api/auth"
	oauth2api "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/api/oauth2"
	appsvc "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/app"
	oauth2svc "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/oauth2"
	shopSvc "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/shop"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func main() {
	initSQLiteAndSeedData()

	appsvc.Init()
	oauth2svc.Init()
	shopSvc.Init()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	oauthApp := r.Group("/oauth/app")
	oauthApp.POST("/token", oauth2api.Token)

	oauthApp.Use(authapi.RequireOAuthAuth)
	oauthApp.GET("/authorize", oauth2api.AuthorizePage)
	oauthApp.POST("/create_auth_session", oauth2api.CreateAuthSession)
	oauthApp.POST("/authorize", oauth2api.Authorize)
	oauthApp.GET("/redirect", oauth2api.RedirectPage)

	r.GET("/auth/login", authapi.LoginPage)
	r.POST("/auth/login", authapi.Login)

	r.GET("/apps/page", authapi.RequirePageAuth, appapi.ListPage)
	r.GET("/apps", authapi.RequireAPIAuth, appapi.List)
	r.GET("/apps/:id", appapi.GetByID)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3360"
	}

	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

func initSQLiteAndSeedData() {
	dsn := os.Getenv("AUTH_SERVER_DB_DSN")
	if len(dsn) == 0 {
		// mem
		dsn = "sqlite://file::memory:?cache=shared"
	}

	driver.InitDB(config.DB{DSN: dsn})
	db := driver.GetDB()

	if err := db.AutoMigrate(&object.App{}, &object.ShopApp{}, &object.Shop{}, &object.Account{}); err != nil {
		panic(fmt.Sprintf("failed to migrate auth-server sqlite tables: %v", err))
	}

	apps := []object.App{
		{
			ID:              "app_l0g6nfq800002c86amnkbcaa",
			Name:            "Demo App",
			Secret:          "mock_secret",
			InstallationURL: "http://localhost:9094",
			RedirectURL:     "http://localhost:9094/oauth",
			HomePageURL:     "http://localhost:9094/home",
		},
		// {
		// 	ID:              "app_demo_2",
		// 	Name:            "Demo App 2",
		// 	Secret:          "mock_secret_2",
		// 	InstallationURL: "http://localhost:9094",
		// 	RedirectURL:     "http://localhost:9094/oauth",
		// 	HomePageURL:     "http://localhost:9094/home",
		// },
	}

	for _, app := range apps {
		if err := db.Where("id = ?", app.ID).FirstOrCreate(&app).Error; err != nil {
			panic(fmt.Sprintf("failed to seed app data: %v", err))
		}
	}

	shopApps := []object.ShopApp{
		{AppID: "app_l0g6nfq800002c86amnkbcaa", ShopID: "shop_demo", Status: "installed"},
		// {AppID: "app_demo_2", ShopID: "shop_demo", Status: "installed"},
	}

	for _, shopApp := range shopApps {
		if err := db.Where("app_id = ? AND shop_id = ?", shopApp.AppID, shopApp.ShopID).FirstOrCreate(&shopApp).Error; err != nil {
			panic(fmt.Sprintf("failed to seed shop_app data: %v", err))
		}
	}

	shops := []object.Shop{
		{ID: "shop_demo", Name: "Demo Shop", Status: "active"},
	}

	for _, shop := range shops {
		if err := db.Where("id = ?", shop.ID).FirstOrCreate(&shop).Error; err != nil {
			panic(fmt.Sprintf("failed to seed shop data: %v", err))
		}
	}

	accounts := []object.Account{
		{ID: "acc_demo", Email: "admin@shop.com", Password: "admin"},
	}

	for _, account := range accounts {
		if err := db.Where("id = ?", account.ID).FirstOrCreate(&account).Error; err != nil {
			panic(fmt.Sprintf("failed to seed account data: %v", err))
		}
	}
}
