package app

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _app App
var _initAppOnce sync.Once

type (
	App interface {
		List(ctx context.Context) ([]*object.App, error)
		CreateAuthSession(ctx context.Context, authInfo *jsonmodel.AuthInfo, payload *jsonmodel.OAuthPayload) (string, error)
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

参考 https://github.com/go-oauth2/oauth2/tree/master/example 编写代码，在官方demo基础上修改