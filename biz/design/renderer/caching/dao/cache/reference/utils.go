package reference

import "time"

const defaultTTL = 10 * time.Minute

func productRefKey(productID string) string {
	return "reference:product:" + productID
}

func productVariantRefKey(variantID string) string {
	return "reference:product_variant:" + variantID
}
