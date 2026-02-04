package image

import (
	"fmt"
	"time"
)

const defaultTTL = 10 * time.Minute

func imageKey(imageID string) string {
	return fmt.Sprintf("image:%s", imageID)
}
