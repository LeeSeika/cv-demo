package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	authapi "github.com/leeseika/cv-demo/biz/design/oauth2/resource-server/api/auth"
	orderapi "github.com/leeseika/cv-demo/biz/design/oauth2/resource-server/api/order"
	ordersvc "github.com/leeseika/cv-demo/biz/design/oauth2/resource-server/service/order"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func main() {
	initSQLiteAndSeedData()
	ordersvc.Init()

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.Use(authapi.RequireAuth)
	api.GET("/orders", orderapi.List)

	port := os.Getenv("RESOURCE_SERVER_PORT")
	if len(port) == 0 {
		port = "3370"
	}

	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

func initSQLiteAndSeedData() {
	dsn := os.Getenv("RESOURCE_SERVER_DB_DSN")
	if len(dsn) == 0 {
		dsn = "sqlite://resource_server.db"
	}

	driver.InitDB(config.DB{DSN: dsn})
	db := driver.GetDB()

	if err := db.AutoMigrate(&object.Order{}); err != nil {
		panic(fmt.Sprintf("failed to migrate resource-server tables: %v", err))
	}

	orders := []object.Order{
		{ID: "ord_001", ShopID: "shop_demo", OrderNo: "1001", Status: "paid", TotalAmount: 12900},
		{ID: "ord_002", ShopID: "shop_demo", OrderNo: "1002", Status: "pending", TotalAmount: 9800},
		{ID: "ord_003", ShopID: "shop_other", OrderNo: "2001", Status: "paid", TotalAmount: 18800},
	}

	for _, order := range orders {
		if err := db.Where("id = ?", order.ID).FirstOrCreate(&order).Error; err != nil {
			panic(fmt.Sprintf("failed to seed order data: %v", err))
		}
	}
}
