package account

import (
	"context"
	"sync"

	accountdao "github.com/leeseika/cv-demo/biz/design/proj-structure/crdb-tx/dao/account"
	shopdao "github.com/leeseika/cv-demo/biz/design/proj-structure/crdb-tx/dao/shop"
)

var _account *account
var _initAccountOnce sync.Once

type (
	Account interface {
		AssignShop(ctx context.Context, accountID string, shopID string) error
	}

	account struct {
		accountDAO accountdao.Account
		shopDAO    shopdao.Shop
	}
)

func Init() {
	_initAccountOnce.Do(func() {
		_account = &account{
			accountDAO: accountdao.Get(),
			shopDAO:    shopdao.Get(),
		}
	})
}

func Get() Account {
	return _account
}
