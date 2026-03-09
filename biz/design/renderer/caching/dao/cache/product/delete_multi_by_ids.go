package product

import "context"

func (p *product) DeleteMultiByIDs(ctx context.Context, productIDs []string) error {
	return p.kvCache.DeleteMulti(ctx, productIDs)
}
