package shop

import "context"

func (p *shop) SetNil(ctx context.Context, shopID string) error {
	key := shopKey(shopID)

	return p.kvCache.SetEmptyValuePlaceholder(ctx, key)
}
