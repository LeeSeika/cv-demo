package shop

import (
	"context"

	shopCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/shop"
	shopDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/shop"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _shop *shop

type (
	Shop interface {
		GetByID(ctx context.Context, shopID string) (*cache.Shop, error)
	}
	shop struct {
		shopDAO      shopDAO.Shop
		shopCacheDAO shopCacheDAO.Shop
	}
)
