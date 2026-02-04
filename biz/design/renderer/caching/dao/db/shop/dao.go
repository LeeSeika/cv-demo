package shop

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _shop Shop
var _initShopOnce sync.Once

type (
	Shop interface {
		GetByID(ctx context.Context, shopID string) (*object.Shop, error)
	}

	shop struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) {
	_initShopOnce.Do(func() {
		_shop = &shop{
			db: db,
		}
	})
}

func Get() Shop {
	return _shop
}
