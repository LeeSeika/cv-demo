package product

import (
	"context"

	productCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/product"
	productvariantCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/product-variant"
	productDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product"
	productvariantDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/product-variant"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _product Product

type (
	Product interface {
		GetProductByID(ctx context.Context, productID string) (*cache.Product, error)
		GetProductVariantsByIDs(ctx context.Context, productVariantIDs []string) (map[string]*cache.ProductVariant, error)
	}

	product struct {
		productDAO             productDAO.Product
		productVariantDAO      productvariantDAO.ProductVariant
		productCacheDAO        productCacheDAO.Product
		productVariantCacheDAO productvariantCacheDAO.ProductVariant
	}
)

func Init(
	productDAO productDAO.Product,
	productCacheDAO productCacheDAO.Product,
	productVariantDAO productvariantDAO.ProductVariant,
	productVariantCacheDAO productvariantCacheDAO.ProductVariant,
) {
	_product = &product{
		productDAO:             productDAO,
		productCacheDAO:        productCacheDAO,
		productVariantDAO:      productVariantDAO,
		productVariantCacheDAO: productVariantCacheDAO,
	}
}

func Get() Product {
	return _product
}
