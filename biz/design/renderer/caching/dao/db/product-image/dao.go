package productimage

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var p ProductImage = &productImage{}
var _initProductImageOnce sync.Once

type (
	ProductImage interface {
		GetImageIDsByProductIDs(ctx context.Context, productIDs []string) (map[string][]*object.ProductImage, error)
	}

	productImage struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) ProductImage {
	_initProductImageOnce.Do(func() {
		p = &productImage{
			db: db,
		}
	})
	return p
}

func Get() ProductImage {
	return p
}
