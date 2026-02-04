package product

import (
	"context"

	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/leeseika/cv-demo/pkg/errs"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (p *product) GetProductByID(ctx context.Context, productID string) (*cache.Product, error) {
	productCache, err := p.productCacheDAO.GetByID(ctx, productID)
	if err == nil {
		return productCache, nil
	}
	if errs.IsKVError(err, kvcache.ErrKeyNotFound) {
		return nil, err
	}

	productObj, err := p.productDAO.GetByID(ctx, productID)
	if err != nil {
		if errs.IsDBError(err, gorm.ErrRecordNotFound) {
			p.productCacheDAO.SetNil(ctx, productID)
		}
		return nil, err
	}

	productCache = cache.ProductFromObject(productObj)
	err = p.productCacheDAO.Set(ctx, productID, productCache)
	if err != nil {
		log.Err(err).Msg("failed to set product cache")
	}

	return productCache, nil
}
