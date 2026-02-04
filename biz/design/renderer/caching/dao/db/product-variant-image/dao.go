package productvariantimage

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var p ProductVariantImage = &productVariantImage{}
var _initProductVariantImageOnce sync.Once

type (
	ProductVariantImage interface {
		GetImageIDsByProductVariantIDs(ctx context.Context, productVariantIDs []string) (map[string][]*object.ProductVariantImage, error)
	}

	productVariantImage struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) ProductVariantImage {
	_initProductVariantImageOnce.Do(func() {
		p = &productVariantImage{
			db: db,
		}
	})
	return p
}

func Get() ProductVariantImage {
	return p
}
