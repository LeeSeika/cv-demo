package productvariant

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (p *productVariant) GetByIDs(ctx context.Context, productVariantIDs []string) ([]*object.ProductVariant, error) {
	db := p.db.WithContext(ctx)

	var productVariants []*object.ProductVariant
	res := db.Model(object.ProductVariant{}).
		Where("id IN ?", productVariantIDs).
		Find(&productVariants)
	if res.Error != nil {
		return nil, res.Error
	}

	return productVariants, nil
}
