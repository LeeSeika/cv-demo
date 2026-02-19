package aggregation

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/dto"
)

var _aggregation Aggregation
var _initAggregationOnce sync.Once

type (
	Aggregation interface {
		GetProductDetail(ctx context.Context, productID string) (*dto.ProductDetail, error)
		SetProductDetail(ctx context.Context, productID string, productDetail *dto.ProductDetail) error
	}

	aggregation struct {
		kvCache driver.KVCacheProvider
	}
)

func Init() {
	_initAggregationOnce.Do(func() {
		_aggregation = &aggregation{
			kvCache: driver.GetKVCacheProvider(),
		}
	})
}

func Get() Aggregation {
	return _aggregation
}
