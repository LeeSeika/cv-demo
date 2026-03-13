package account

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _account *account
var _initAccountOnce sync.Once

type (
	Account interface {
		CreateAccount(ctx context.Context, obj *object.Account) error
		GetAccountByID(ctx context.Context, id string) (*object.Account, error)
		UpdateAccount(ctx context.Context, id string, obj *object.Account) error
	}

	account struct {
		db *gorm.DB
	}
)

func Init() {
	_initAccountOnce.Do(func() {
		_account = &account{
			db: driver.GetDB(),
		}
	})
}

func Get() Account {
	return _account
}
