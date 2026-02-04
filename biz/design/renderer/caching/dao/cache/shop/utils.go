package shop

import (
	"fmt"
	"time"
)

const defaultTTL = 10 * time.Minute

func shopKey(shopID string) string {
	return fmt.Sprintf("shop:%s", shopID)
}
