package oauth2

import (
	"context"
	"sync"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

var _oauth2 Oauth2
var _initOauth2Once sync.Once

type (
	Oauth2 interface {
		AuthorizePage(ctx context.Context) (string, error)
		Authorize(ctx context.Context) error
	}

	oauth2 struct {
		srv *server.Server
	}
)

func Init() {
	_initOauth2Once.Do(func() {
		manager := manage.NewDefaultManager()
		// token memory store
		manager.MustTokenStorage(store.NewMemoryTokenStore())
		srv := server.NewDefaultServer(manager)
		_oauth2 = &oauth2{
			srv: srv,
		}
	})
}

func Get() Oauth2 {
	return _oauth2
}
