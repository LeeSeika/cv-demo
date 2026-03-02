package main

import (
	"fmt"
	"os"

	api "github.com/leeseika/cv-demo/biz/design/2fa/api"
	accountsvc "github.com/leeseika/cv-demo/biz/design/2fa/service/account"
	authsvc "github.com/leeseika/cv-demo/biz/design/2fa/service/auth"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func main() {
	initDB()
	authsvc.Init()
	accountsvc.Init()

	port := os.Getenv("TWO_FA_PORT")
	if len(port) == 0 {
		port = "9100"
	}

	server := api.NewServer(":" + port)
	if err := server.Start(); err != nil {
		panic(err)
	}
}

func initDB() {
	dsn := os.Getenv("TWO_FA_DB_DSN")
	if len(dsn) == 0 {
		dsn = "sqlite://file::memory:?cache=shared"
	}

	driver.InitDB(config.DB{DSN: dsn})
	db := driver.GetDB()

	if err := db.AutoMigrate(&object.Account{}); err != nil {
		panic(fmt.Sprintf("failed to migrate 2fa tables: %v", err))
	}
}
