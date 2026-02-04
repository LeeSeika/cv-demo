package product

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (p *product) GetByID(ctx context.Context, productID string) (*cache.Product, error) {
	key := productKey(productID)

	b, err := p.kvCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var productCache cache.Product
	err = json.Unmarshal(b, &productCache)
	if err != nil {
		return nil, err
	}

	return &productCache, nil
}
