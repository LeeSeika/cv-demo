package dto

import "github.com/leeseika/cv-demo/pkg/model/cache"

type ProductDetail struct {
	cache.Product
	Variants []*ProductVariantDetail `json:"variants"`
	Images   []*cache.Image          `json:"images"`
}

func BuildProductDetail(
	product cache.Product,
	variants []*ProductVariantDetail,
	images []*cache.Image,
) *ProductDetail {
	return &ProductDetail{
		Product:  product,
		Variants: variants,
		Images:   images,
	}
}
