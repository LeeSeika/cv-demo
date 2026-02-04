package product

import (
	"fmt"
	"time"
)

const defaultTTL = 10 * time.Minute

func productKey(productID string) string {
	return fmt.Sprintf("product:%s", productID)
}
