package shop

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (s *shop) GetByID(ctx context.Context, shopID string) (*object.Shop, error) {
	db := s.db.WithContext(ctx)

	var shop object.Shop
	db.Model(object.Shop{}).
		Where("id = ?", shopID).
		First(&shop)
	if db.Error != nil {
		return nil, db.Error
	}

	return &shop, nil
}
