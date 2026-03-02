package aggregation

import (
	"context"

	aggregationCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/aggregation"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"golang.org/x/sync/singleflight"
)

var _aggregation Aggregation

type (
	Aggregation interface {
		GetProductDetail(ctx context.Context, productID string) (*dto.ProductDetail, error)
	}

	aggregation struct {
		aggregationCacheDAO aggregationCacheDAO.Aggregation
		singleFlightGroup   singleflight.Group
	}
)

func Init(aggregationCacheDAO aggregationCacheDAO.Aggregation) {
	_aggregation = &aggregation{
		aggregationCacheDAO: aggregationCacheDAO,
	}
}

func Get() Aggregation {
	return _aggregation
}
