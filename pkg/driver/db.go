package driver

import (
	"sync"

	"github.com/leeseika/cv-demo/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	_db         *gorm.DB
	_initDBOnce sync.Once
)

func InitDB(conf config.DB) {
	_initDBOnce.Do(func() {
		db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{})
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
		_db = db
	})
}

func GetDB() *gorm.DB {
	return _db
}
