package image

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (i *image) GetByID(ctx context.Context, imageID string) (*cache.Image, error) {
	key := imageKey(imageID)

	b, err := i.kvCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var imageCache cache.Image
	err = json.Unmarshal(b, &imageCache)
	if err != nil {
		return nil, err
	}

	return &imageCache, nil
}
