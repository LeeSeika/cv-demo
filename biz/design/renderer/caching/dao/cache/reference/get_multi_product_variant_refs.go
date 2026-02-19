package reference

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (r *reference) GetMultiProductVariantRefs(ctx context.Context, variantIDs []string) (map[string]*cache.ProductVariantReference, error) {
	keys := make([]string, 0, len(variantIDs))
	result := make(map[string]*cache.ProductVariantReference, len(variantIDs))
	for _, variantID := range variantIDs {
		key := productVariantRefKey(variantID)
		keys = append(keys, key)
	}

	values, err := r.kvCache.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	for _, value := range values {
		if value == nil {
			continue
		}

		var variantRefCache cache.ProductVariantReference
		err = json.Unmarshal(value, &variantRefCache)
		if err != nil {
			continue
		}

		result[variantRefCache.ID] = &variantRefCache
	}

	return result, nil
}
