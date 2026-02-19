package reference

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _reference Reference
var _initReferenceOnce sync.Once

type (
	Reference interface {
		GetProductRef(ctx context.Context, productID string) (*cache.ProductReference, error)
		GetMultiProductVariantRefs(ctx context.Context, variantIDs []string) (map[string]*cache.ProductVariantReference, error)
		SetProductRef(ctx context.Context, productID string, productRef *cache.ProductReference) error
		SetMultiProductVariantRefs(ctx context.Context, variantRefs map[string]*cache.ProductVariantReference) error
	}

	reference struct {
		kvCache driver.KVCacheProvider
	}
)

func Init() {
	_initReferenceOnce.Do(func() {
		_reference = &reference{
			kvCache: driver.GetKVCacheProvider(),
		}
	})
}

func Get() Reference {
	return _reference
}
