package product

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (p *product) GetByID(ctx context.Context, productID string) (*object.Product, error) {
	db := p.db.WithContext(ctx)

	var product object.Product
	db.Model(object.Product{}).
		Where("id = ?", productID).
		First(&product)
	if db.Error != nil {
		return nil, db.Error
	}

	return &product, nil
}
