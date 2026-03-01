package order

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (o *order) ListByShopID(ctx context.Context, shopID string) ([]*object.Order, error) {
	var orders []*object.Order
	err := o.db.WithContext(ctx).Where("shop_id = ?", shopID).Order("created_at desc").Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}
