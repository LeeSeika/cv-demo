package productvariant

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (pv *productVariant) GetByIDs(ctx context.Context, productVariantIDs []string) (map[string]*cache.ProductVariant, error) {
	if len(productVariantIDs) == 0 {
		return map[string]*cache.ProductVariant{}, nil
	}

	keys := make([]string, 0, len(productVariantIDs))
	for _, productVariantID := range productVariantIDs {
		keys = append(keys, productVariantKey(productVariantID))
	}

	bMap, err := pv.kvCache.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	var lastErr error

	productVariantCaches := make(map[string]*cache.ProductVariant, len(keys))
	for _, productVariantID := range productVariantIDs {
		key := productVariantKey(productVariantID)
		b, ok := bMap[key]
		if !ok {
			continue
		}
		var productVariantCache cache.ProductVariant
		err := json.Unmarshal(b, &productVariantCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal productVariant cache json")
			lastErr = err
		}
		productVariantCaches[productVariantID] = &productVariantCache
	}

	return productVariantCaches, lastErr
}
