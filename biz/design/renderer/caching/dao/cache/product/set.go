package product

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (p *product) Set(ctx context.Context, productID string, product *cache.Product) error {
	if product == nil || productID == "" {
		return nil
	}

	key := productKey(productID)

	b, err := json.Marshal(product)
	if err != nil {
		return err
	}

	err = p.kvCache.Set(ctx, key, b, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
