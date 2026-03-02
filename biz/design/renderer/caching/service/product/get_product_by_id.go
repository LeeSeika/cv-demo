package product

import (
	"context"
	"fmt"
	"time"

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
	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		return nil, err
	}

	resultCh := p.singleFlightGroup.DoChan(fmt.Sprintf("product:%s", productID), func() (interface{}, error) {
		cached, cacheErr := p.productCacheDAO.GetByID(ctx, productID)
		if cacheErr == nil {
			return cached, nil
		}
		if errs.IsKVCacheError(cacheErr, kvcache.ErrKeyNotFound) {
			return nil, cacheErr
		}

		productObj, dbErr := p.productDAO.GetByID(ctx, productID)
		if dbErr != nil {
			if errs.IsDBError(dbErr, gorm.ErrRecordNotFound) {
				p.productCacheDAO.SetNil(ctx, productID)
			}
			return nil, dbErr
		}

		rebuilt := cache.ProductFromObject(productObj)
		setErr := p.productCacheDAO.Set(ctx, productID, rebuilt)
		if setErr != nil {
			log.Err(setErr).Msg("failed to set product cache")
		}

		return rebuilt, nil
	})

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout.C:
		return nil, context.DeadlineExceeded
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*cache.Product), nil
	}
}
