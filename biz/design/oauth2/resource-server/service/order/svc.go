package order

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _order Order
var _initOrderOnce sync.Once

type (
	Order interface {
		ListByShopID(ctx context.Context, shopID string) ([]*object.Order, error)
	}

	order struct {
		db *gorm.DB
	}
)

func Init() {
	_initOrderOnce.Do(func() {
		_order = &order{db: driver.GetDB()}
	})
}

func Get() Order {
	return _order
}
