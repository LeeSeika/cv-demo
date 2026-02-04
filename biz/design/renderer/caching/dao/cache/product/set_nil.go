package product

import "context"

func (p *product) SetNil(ctx context.Context, productID string) error {
	key := productKey(productID)

	return p.kvCache.SetEmptyValuePlaceholder(ctx, key)
}
