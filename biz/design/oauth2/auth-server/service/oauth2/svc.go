package oauth2

import (
	"context"
	"sync"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"gorm.io/gorm"
)

var _oauth2 OAuth2
var _srv *server.Server
var _initOAuth2Once sync.Once

type (
	OAuth2 interface {
		AuthorizePage(ctx context.Context, payload dto.OAuthPayload) (string, error)
		OAuth2Server() *server.Server
		// CreateAuthSession 是在授权过程中创建一个临时的授权会话（由 js 在渲染安装页面时自动调用），保证进入 auth 页面和点击授权按钮的是同一个用户，防止 CSRF 攻击
		// 这并不是一个标准的 OAuth2 流程，而是我们为了安全性在授权流程中添加的一个额外步骤
		// 在点击安装后，发请求到 oauth/app/authorize 的时候，就会校验这个 token
		CreateAuthSession(ctx context.Context, authInfo jsonmodel.AuthInfo, payload dto.OAuthPayload) (string, string, error)
		ValidateCSRFToken(ctx context.Context, csrfToken string, hashed string, authorizeRequest *server.AuthorizeRequest) error
	}

	oauth2 struct {
		db *gorm.DB
	}
)

func Init() {
	_initOAuth2Once.Do(func() {
		_oauth2 = &oauth2{
			db: driver.GetDB(),
		}

		manager := manage.NewDefaultManager()

		// client store
		clientStore := newAppInfoStore(_oauth2)
		manager.MapClientStorage(clientStore)
		// token store
		manager.MustTokenStorage(newAppAccessTokenStore(_oauth2))
		srv := server.NewDefaultServer(manager)
		srv.SetUserAuthorizationHandler(authHandler)

		_srv = srv
	})
}

func Get() OAuth2 {
	return _oauth2
}
