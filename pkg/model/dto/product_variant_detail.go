package dto

import "github.com/leeseika/cv-demo/pkg/model/cache"

type ProductVariantDetail struct {
	cache.ProductVariant
	Price        variantPrice   `json:"price"`
	ProductTitle string         `json:"product_title"`
	Images       []*cache.Image `json:"image"`
}

type variantPrice struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

func BuildProductVariantDetail(
	variant cache.ProductVariant,
	productTitle string,
	currencyCode string,
	images []*cache.Image,
) *ProductVariantDetail {
	return &ProductVariantDetail{
		ProductVariant: variant,
		Price: variantPrice{
			Amount:       variant.Price,
			CurrencyCode: currencyCode,
		},
		ProductTitle: productTitle,
		Images:       images,
	}
}
