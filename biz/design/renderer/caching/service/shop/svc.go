package shop

import (
	"context"
	"sync"

	shopCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/shop"
	shopDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/shop"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _shop *shop
var _initShopOnce sync.Once

type (
	Shop interface {
		GetByID(ctx context.Context, shopID string) (*cache.Shop, error)
	}
	shop struct {
		shopDAO      shopDAO.Shop
		shopCacheDAO shopCacheDAO.Shop
	}
)

func Init() {
	_initShopOnce.Do(func() {
		_shop = &shop{
			shopDAO:      shopDAO.Get(),
			shopCacheDAO: shopCacheDAO.Get(),
		}
	})
}

func Get() Shop {
	return _shop
}
