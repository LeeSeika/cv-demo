package image

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (i *image) GetByIDs(ctx context.Context, imageIDs []string) (map[string]*cache.Image, error) {
	if len(imageIDs) == 0 {
		return map[string]*cache.Image{}, nil
	}

	keys := make([]string, 0, len(imageIDs))
	for _, imageID := range imageIDs {
		keys = append(keys, imageKey(imageID))
	}

	bMap, err := i.kvCache.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	var lastErr error

	imageCaches := make(map[string]*cache.Image, len(keys))
	for _, imageID := range imageIDs {
		key := imageKey(imageID)
		b, ok := bMap[key]
		if !ok {
			continue
		}
		var imageCache cache.Image
		err := json.Unmarshal(b, &imageCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal image cache json")
			lastErr = err
		}
		imageCaches[imageID] = &imageCache
	}

	return imageCaches, lastErr
}
