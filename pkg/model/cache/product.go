package cache

import (
	"time"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

type Product struct {
	ID          string                     `json:"id"`
	ShopID      string                     `json:"shop_id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Options     []*jsonmodel.ProductOption `json:"options"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	CreatedAt   time.Time                  `json:"created_at"`
}

func ProductFromObject(productObj *object.Product) *Product {
	return &Product{
		ID:          productObj.ID,
		ShopID:      productObj.ShopID,
		Title:       productObj.Title,
		Description: productObj.Description,
		Options:     productObj.Options,
		UpdatedAt:   productObj.UpdatedAt,
		CreatedAt:   productObj.CreatedAt,
	}
}
