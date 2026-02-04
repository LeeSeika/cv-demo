package product

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _product Product
var _initProductOnce sync.Once

type (
	Product interface {
		GetByID(ctx context.Context, productID string) (*object.Product, error)
	}

	product struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) {
	_initProductOnce.Do(func() {
		_product = &product{
			db: db,
		}
	})
}

func Get() Product {
	return _product
}
