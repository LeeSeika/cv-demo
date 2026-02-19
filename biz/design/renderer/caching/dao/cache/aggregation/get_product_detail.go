package aggregation

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (a *aggregation) GetProductDetail(ctx context.Context, productID string) (*dto.ProductDetail, error) {
	key := productDetailKey(productID)

	b, err := a.kvCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var productDetailCache dto.ProductDetail
	err = json.Unmarshal(b, &productDetailCache)
	if err != nil {
		return nil, err
	}

	return &productDetailCache, nil
}
