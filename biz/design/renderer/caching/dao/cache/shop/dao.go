package shop

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _shop Shop
var _initShopOnce sync.Once

type (
	Shop interface {
		GetByID(ctx context.Context, shopID string) (*cache.Shop, error)
		Set(ctx context.Context, shopID string, shop *cache.Shop) error
		SetNil(ctx context.Context, shopID string) error
	}

	shop struct {
		kvCache driver.KVCacheProvider
	}
)

func Init(kvCache driver.KVCacheProvider) {
	_initShopOnce.Do(func() {
		_shop = &shop{
			kvCache: kvCache,
		}
	})
}

func Get() Shop {
	return _shop
}
