package reference

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (r *reference) GetProductRef(ctx context.Context, productID string) (*cache.ProductReference, error) {
	key := productRefKey(productID)

	b, err := r.kvCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var productRefCache cache.ProductReference
	err = json.Unmarshal(b, &productRefCache)
	if err != nil {
		return nil, err
	}

	return &productRefCache, nil
}
