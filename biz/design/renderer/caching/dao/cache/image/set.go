package image

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (i *image) Set(ctx context.Context, imageID string, image *cache.Image) error {
	if image == nil || imageID == "" {
		return nil
	}

	key := imageKey(imageID)

	b, err := json.Marshal(image)
	if err != nil {
		return err
	}

	err = i.kvCache.Set(ctx, key, b, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
