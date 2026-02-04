package image

import (
	"context"
	"maps"

	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (i *image) GetImagesByIDs(ctx context.Context, imageIDs []string) (map[string]*cache.Image, error) {
	if len(imageIDs) < 1 {
		return map[string]*cache.Image{}, nil
	}

	imageCaches, err := i.imageCacheDAO.GetByIDs(ctx, imageIDs)
	if err != nil {
		log.Err(err).Msg("failed to get image caches")
	}

	missingImageIDs := make([]string, 0, len(imageIDs))
	for _, imageID := range imageIDs {
		_, ok := imageCaches[imageID]
		if !ok {
			missingImageIDs = append(missingImageIDs, imageID)
		}
	}

	if len(missingImageIDs) > 0 {
		imageObjs, err := i.imageDAO.GetByIDs(ctx, missingImageIDs)
		if err != nil {
			log.Err(err).Msg("failed to get image objects from database")
			return imageCaches, err
		}

		rebuildImageCaches := make(map[string]*cache.Image, len(missingImageIDs))
		for _, imageObj := range imageObjs {
			imageCache := cache.ImageFromObject(imageObj)
			rebuildImageCaches[imageObj.ID] = imageCache
		}

		err = i.imageCacheDAO.SetMulti(ctx, rebuildImageCaches)
		if err != nil {
			log.Err(err).Msg("failed to rebuild image caches")
		}

		maps.Copy(imageCaches, rebuildImageCaches)
	}

	return imageCaches, nil
}
