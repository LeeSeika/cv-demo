package productvariant

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _productVariant ProductVariant
var _initProductVariantOnce sync.Once

type (
	ProductVariant interface {
		GetByIDs(ctx context.Context, productVariantIDs []string) (map[string]*cache.ProductVariant, error)
		SetMulti(ctx context.Context, productVariants map[string]*cache.ProductVariant) error
	}

	productVariant struct {
		kvCache driver.KVCacheProvider
	}
)

func Init(kvCache driver.KVCacheProvider) {
	_initProductVariantOnce.Do(func() {
		_productVariant = &productVariant{
			kvCache: kvCache,
		}
	})
}

func Get() ProductVariant {
	return _productVariant
}
