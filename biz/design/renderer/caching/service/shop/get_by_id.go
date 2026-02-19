package shop

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (s *shop) GetByID(ctx context.Context, shopID string) (*cache.Shop, error) {
	// try to get from cache
	shopCache, err := s.shopCacheDAO.GetByID(ctx, shopID)
	if err == nil {
		return shopCache, nil
	}

	// query from db
	shopObject, err := s.shopDAO.GetByID(ctx, shopID)
	if err != nil {
		return nil, err
	}

	shopCache = cache.BuildShop(
		shopID,
		shopObject.Name,
		shopCache.CurrencyCode,
	)

	// set cache
	err = s.shopCacheDAO.Set(ctx, shopID, shopCache)
	if err != nil {
		log.Err(err).Msg("failed to set shop cache")
	}

	return shopCache, nil
}
