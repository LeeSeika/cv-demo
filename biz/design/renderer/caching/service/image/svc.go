package image

import (
	"context"

	imageCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/cache/image"
	imageDAO "github.com/leeseika/cv-demo/biz/design/renderer/caching/dao/db/image"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _image Image

type (
	Image interface {
		GetImagesByIDs(ctx context.Context, imageIDs []string) (map[string]*cache.Image, error)
	}

	image struct {
		imageDAO      imageDAO.Image
		imageCacheDAO imageCacheDAO.Image
	}
)

func Init(imageDAO imageDAO.Image, imageCacheDAO imageCacheDAO.Image) {
	_image = &image{
		imageDAO:      imageDAO,
		imageCacheDAO: imageCacheDAO,
	}
}

func Get() Image {
	return _image
}
