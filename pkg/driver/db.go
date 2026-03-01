package driver

import (
	"strings"
	"sync"

	"github.com/leeseika/cv-demo/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	_db         *gorm.DB
	_initDBOnce sync.Once
)

func InitDB(conf config.DB) {
	_initDBOnce.Do(func() {
		var dialector gorm.Dialector

		switch {
		case strings.HasPrefix(conf.DSN, "sqlite://"):
			dialector = sqlite.Open(strings.TrimPrefix(conf.DSN, "sqlite://"))
		case strings.HasPrefix(conf.DSN, "sqlite:"):
			dialector = sqlite.Open(strings.TrimPrefix(conf.DSN, "sqlite:"))
		default:
			dialector = postgres.Open(conf.DSN)
		}

		db, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
		_db = db
	})
}

func GetDB() *gorm.DB {
	return _db
}
