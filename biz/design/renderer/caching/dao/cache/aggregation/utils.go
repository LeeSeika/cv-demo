package aggregation

import "time"

const defaultTTL = 10 * time.Second

func productDetailKey(productID string) string {
	return "aggregation:product_detail:" + productID
}
