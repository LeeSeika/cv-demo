package aggregation

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (a *aggregation) SetProductDetail(ctx context.Context, productID string, productDetail *dto.ProductDetail) error {
	key := productDetailKey(productID)

	b, err := json.Marshal(productDetail)
	if err != nil {
		return err
	}

	err = a.kvCache.Set(ctx, key, b, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
