package reference

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (r *reference) SetMultiProductVariantRefs(ctx context.Context, variantRefs map[string]*cache.ProductVariantReference) error {
	m := make(map[string][]byte, len(variantRefs))

	for variantID, variantRef := range variantRefs {
		key := productVariantRefKey(variantID)
		b, err := json.Marshal(variantRef)
		if err != nil {
			continue
		}
		m[key] = b
	}

	err := r.kvCache.SetMulti(ctx, m, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
