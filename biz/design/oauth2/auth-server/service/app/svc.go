package app

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _app App
var _initAppOnce sync.Once

type (
	App interface {
		List(ctx context.Context) ([]*object.App, error)
		GetByID(ctx context.Context, appID string) (*object.App, error)
	}

	app struct {
		db *gorm.DB
	}
)

func Init() {
	_initAppOnce.Do(func() {
		_app = &app{
			db: driver.GetDB(),
		}
	})
}

func Get() App {
	return _app
}
