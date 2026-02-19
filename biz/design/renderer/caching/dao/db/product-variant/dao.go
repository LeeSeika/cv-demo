package productvariant

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _productVariant ProductVariant
var _initProductVariantOnce sync.Once

type (
	ProductVariant interface {
		GetByIDs(ctx context.Context, productVariantIDs []string) ([]*object.ProductVariant, error)
		GetVariantIDsByProductID(ctx context.Context, productID string) ([]string, error)
	}

	productVariant struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) {
	_initProductVariantOnce.Do(func() {
		_productVariant = &productVariant{
			db: db,
		}
	})
}

func Get() ProductVariant {
	return _productVariant
}
