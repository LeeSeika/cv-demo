package product

import "context"

func (p *product) DeleteCacheByIDs(ctx context.Context, productIDs []string) error {
	err := p.productCacheDAO.DeleteMultiByIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	return nil
}
