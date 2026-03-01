package shop

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (s *shop) VerifyShopAvailability(ctx context.Context, shopID string) (bool, error) {
	var shop object.Shop
	err := s.db.First(&shop, "id = ?", shopID).Error
	if err != nil {
		return false, err
	}

	if shop.Status != "active" {
		return false, nil
	}

	return true, nil
}
