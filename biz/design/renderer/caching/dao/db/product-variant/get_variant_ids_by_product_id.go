package productvariant

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (p *productVariant) GetVariantIDsByProductID(ctx context.Context, productID string) ([]string, error) {
	db := p.db.WithContext(ctx)

	var variantIDs []string
	res := db.Model(object.ProductVariant{}).
		Select("id").
		Where("product_id = ?", productID).
		Find(&variantIDs)
	if res.Error != nil {
		return nil, res.Error
	}

	return variantIDs, nil
}
