package reference

import (
	"context"

	referenceCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/reference"
	productDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product"
	productImageDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product-image"
	productVariantDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product-variant"
	productVariantImageDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product-variant-image"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _reference Reference

type (
	Reference interface {
		GetProductRef(ctx context.Context, productID string) (*cache.ProductReference, error)
		GetMultiProductVariantRefs(ctx context.Context, variantIDs []string) (map[string]*cache.ProductVariantReference, error)
	}

	reference struct {
		referenceCacheDAO      referenceCacheDAO.Reference
		productDAO             productDAO.Product
		productVariantDAO      productVariantDAO.ProductVariant
		productImageDAO        productImageDAO.ProductImage
		productVariantImageDAO productVariantImageDAO.ProductVariantImage
	}
)

func Init() {
	_reference = &reference{
		referenceCacheDAO:      referenceCacheDAO.Get(),
		productDAO:             productDAO.Get(),
		productVariantDAO:      productVariantDAO.Get(),
		productImageDAO:        productImageDAO.Get(),
		productVariantImageDAO: productVariantImageDAO.Get(),
	}
}

func Get() Reference {
	return _reference
}
