package account

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"gorm.io/gorm"
)

var _account *account
var _initAccountOnce sync.Once

type (
	Account interface {
		CheckExists(ctx context.Context, id string) (bool, error)
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
