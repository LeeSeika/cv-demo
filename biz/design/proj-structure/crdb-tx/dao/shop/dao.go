package shop

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"gorm.io/gorm"
)

var _shop *shop
var _initShopOnce sync.Once

type (
	Shop interface {
		UpdateAssignAccount(ctx context.Context, shopID string, accountID string) (int64, error)
	}

	shop struct {
		db *gorm.DB
	}
)

func Init() {
	_initShopOnce.Do(func() {
		_shop = &shop{
			db: driver.GetDB(),
		}
	})
}

func Get() Shop {
	return _shop
}
