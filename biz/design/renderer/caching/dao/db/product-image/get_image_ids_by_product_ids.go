package productimage

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (p *productImage) GetImageIDsByProductIDs(ctx context.Context, productIDs []string) (map[string][]*object.ProductImage, error) {
	result := make(map[string][]*object.ProductImage)
	var productImages []*object.ProductImage
	if err := p.db.WithContext(ctx).
		Where("product_id IN ?", productIDs).
		Find(&productImages).Error; err != nil {
		return nil, err
	}

	for _, pi := range productImages {
		result[pi.ProductID] = append(result[pi.ProductID], pi)
	}

	return result, nil
}
