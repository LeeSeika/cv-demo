package product

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _product Product
var _initProductOnce sync.Once

type (
	Product interface {
		GetByID(ctx context.Context, productID string) (*cache.Product, error)
		Set(ctx context.Context, productID string, product *cache.Product) error
		SetNil(ctx context.Context, productID string) error
		DeleteMultiByIDs(ctx context.Context, productIDs []string) error
	}

	product struct {
		kvCache driver.KVCacheProvider
	}
)

func Init(kvCache driver.KVCacheProvider) {
	_initProductOnce.Do(func() {
		_product = &product{
			kvCache: kvCache,
		}
	})
}

func Get() Product {
	return _product
}
