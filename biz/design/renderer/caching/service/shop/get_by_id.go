package shop

import (
	"context"
	"fmt"
	"time"

	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"github.com/rs/zerolog/log"
)

func (s *shop) GetByID(ctx context.Context, shopID string) (*cache.Shop, error) {
	// try to get from cache
	shopCache, err := s.shopCacheDAO.GetByID(ctx, shopID)
	if err == nil {
		return shopCache, nil
	}
	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		return nil, err
	}

	resultCh := s.singleFlightGroup.DoChan(fmt.Sprintf("shop:%s", shopID), func() (interface{}, error) {
		cached, cacheErr := s.shopCacheDAO.GetByID(ctx, shopID)
		if cacheErr == nil {
			return cached, nil
		}
		if errs.IsKVCacheError(cacheErr, kvcache.ErrKeyNotFound) {
			return nil, cacheErr
		}

		shopObject, dbErr := s.shopDAO.GetByID(ctx, shopID)
		if dbErr != nil {
			return nil, dbErr
		}

		rebuilt := cache.BuildShop(
			shopID,
			shopObject.Name,
			shopObject.CurrencyCode,
		)

		setErr := s.shopCacheDAO.Set(ctx, shopID, rebuilt)
		if setErr != nil {
			log.Err(setErr).Msg("failed to set shop cache")
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
		return result.Val.(*cache.Shop), nil
	}
}
