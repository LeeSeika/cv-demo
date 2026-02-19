package reference

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (r *reference) SetProductRef(ctx context.Context, productID string, ref *cache.ProductReference) error {
	key := productRefKey(productID)

	b, err := json.Marshal(ref)
	if err != nil {
		return err
	}

	return r.kvCache.Set(ctx, key, b, defaultTTL)
}
