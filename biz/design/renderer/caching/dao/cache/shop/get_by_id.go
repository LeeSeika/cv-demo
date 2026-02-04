package shop

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (p *shop) GetByID(ctx context.Context, shopID string) (*cache.Shop, error) {
	key := shopKey(shopID)

	b, err := p.kvCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var shopCache cache.Shop
	err = json.Unmarshal(b, &shopCache)
	if err != nil {
		return nil, err
	}

	return &shopCache, nil
}
