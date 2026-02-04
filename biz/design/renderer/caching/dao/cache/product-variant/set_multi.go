package productvariant

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (pv *productVariant) SetMulti(ctx context.Context, productVariants map[string]*cache.ProductVariant) error {
	if len(productVariants) == 0 {
		return nil
	}

	bMap := make(map[string][]byte, len(productVariants))
	for id, productVariantCache := range productVariants {
		productVariantCacheBytes, err := json.Marshal(productVariantCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to marshal productVariant cache json")
			continue
		}
		bMap[productVariantKey(id)] = productVariantCacheBytes
	}
	err := pv.kvCache.SetMulti(ctx, bMap, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
