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
		// CreateAuthSession 是在授权过程中创建一个临时的授权会话（由 js 在渲染安装页面时自动调用），保证进入 auth 页面和点击授权按钮的是同一个用户，防止 CSRF 攻击
		// 这并不是一个标准的 OAuth2 流程，而是我们为了安全性在授权流程中添加的一个额外步骤
		// 在点击安装后，发请求到 oauth/app/authorize 的时候，就会校验这个 token
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