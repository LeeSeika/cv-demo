package image

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (i *image) SetMulti(ctx context.Context, images map[string]*cache.Image) error {
	if len(images) == 0 {
		return nil
	}

	bMap := make(map[string][]byte, len(images))
	for id, imageCache := range images {
		imageCacheBytes, err := json.Marshal(imageCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to marshal image cache json")
			continue
		}
		bMap[imageKey(id)] = imageCacheBytes
	}
	err := i.kvCache.SetMulti(ctx, bMap, defaultTTL)
	if err != nil {
		return err
	}

	return nil
}
