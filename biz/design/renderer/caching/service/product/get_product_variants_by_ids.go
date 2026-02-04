package product

import (
	"context"
	"maps"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (p *product) GetProductVariantsByIDs(ctx context.Context, productVariantIDs []string) (map[string]*cache.ProductVariant, error) {
	if len(productVariantIDs) == 0 {
		return make(map[string]*cache.ProductVariant), nil
	}

	productVariantCaches, err := p.productVariantCacheDAO.GetByIDs(ctx, productVariantIDs)
	if err != nil {
		log.Err(err).Msg("failed to get product variants from cache")
	}

	var missingProductVariantIDs []string
	for _, productVariantID := range productVariantIDs {
		if _, exists := productVariantCaches[productVariantID]; !exists {
			missingProductVariantIDs = append(missingProductVariantIDs, productVariantID)
		}
	}

	if len(missingProductVariantIDs) > 0 {
		productVariantObjs, err := p.productVariantDAO.GetByIDs(ctx, missingProductVariantIDs)
		if err != nil {
			log.Err(err).Msg("failed to get product variants from database")
			return productVariantCaches, err
		}

		rebuildProductVariantCaches := make(map[string]*cache.ProductVariant, len(productVariantObjs))
		for _, productVariantObj := range productVariantObjs {
			productVariantCache := cache.ProductVariantFromObject(productVariantObj)
			rebuildProductVariantCaches[productVariantObj.ID] = productVariantCache
		}

		err = p.productVariantCacheDAO.SetMulti(ctx, rebuildProductVariantCaches)
		if err != nil {
			log.Err(err).Msg("failed to rebuild product variant caches")
		}

		maps.Copy(productVariantCaches, rebuildProductVariantCaches)
	}

	return productVariantCaches, nil
}
