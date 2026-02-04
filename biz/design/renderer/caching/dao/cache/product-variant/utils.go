package productvariant

import (
	"fmt"
	"time"
)

const defaultTTL = 10 * time.Minute

func productVariantKey(productVariantID string) string {
	return fmt.Sprintf("productVariant:%s", productVariantID)
}
