package productvariantimage

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (p *productVariantImage) GetImageIDsByProductVariantIDs(ctx context.Context, productVariantIDs []string) (map[string][]*object.ProductVariantImage, error) {
	result := make(map[string][]*object.ProductVariantImage)
	var productVariantImages []*object.ProductVariantImage
	if err := p.db.WithContext(ctx).
		Where("product_variant_id IN ?", productVariantIDs).
		Find(&productVariantImages).Error; err != nil {
		return nil, err
	}

	for _, pvi := range productVariantImages {
		result[pvi.ProductVariantID] = append(result[pvi.ProductVariantID], pvi)
	}

	return result, nil
}
