package shop

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (p *shop) Set(ctx context.Context, shopID string, shop *cache.Shop) error {
	if shop == nil || shopID == "" {
		return nil
	}

	key := shopKey(shopID)

	b, err := json.Marshal(shop)
	if err != nil {
		return err
	}

	err = p.kvCache.Set(ctx, key, b, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
