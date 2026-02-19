package cache

import (
	"fmt"
	"time"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

type ProductVariant struct {
	ID              string                      `json:"id"`
	Price           string                      `json:"price"`
	SelectedOptions []*jsonmodel.SelectedOption `json:"selected_options"`
	SKU             string                      `json:"sku"`
	Title           string                      `json:"title"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	ProductID       string                      `json:"product_id"`
}

func ProductVariantFromObject(variantObj *object.ProductVariant) *ProductVariant {
	return &ProductVariant{
		ID:              variantObj.ID,
		Price:           fmt.Sprint(variantObj.Price),
		SelectedOptions: variantObj.SelectedOptions,
		CreatedAt:       variantObj.CreatedAt,
		UpdatedAt:       variantObj.UpdatedAt,
		ProductID:       variantObj.ProductID,
	}
}

type ProductVariantReference struct {
	ID       string   `json:"id"`
	ImageIDs []string `json:"image_ids"`
}

func BuildProductVariantReference(
	id string,
	imageIDs []string,
) *ProductVariantReference {
	return &ProductVariantReference{
		ID:       id,
		ImageIDs: imageIDs,
	}
}
