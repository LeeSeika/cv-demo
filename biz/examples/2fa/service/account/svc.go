package account

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _account Account
var _initAccountOnce sync.Once

type (
	Account interface {
		Register(ctx context.Context, email, password string) (string, error)
		Login(ctx context.Context, email, password string) (string, error)
		VerifyOTP(ctx context.Context, accountID string, otp string) (string, error)
		GetAccountInfo(ctx context.Context, accountID string) (*object.Account, error)
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
