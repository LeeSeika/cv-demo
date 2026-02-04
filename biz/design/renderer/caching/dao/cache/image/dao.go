package image

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _image Image
var _initImageOnce sync.Once

type (
	Image interface {
		GetByID(ctx context.Context, imageID string) (*cache.Image, error)
		GetByIDs(ctx context.Context, imageIDs []string) (map[string]*cache.Image, error)
		Set(ctx context.Context, imageID string, image *cache.Image) error
		SetMulti(ctx context.Context, images map[string]*cache.Image) error
		SetNil(ctx context.Context, imageID string) error
	}

	image struct {
		kvCache driver.KVCacheProvider
	}
)

func Init(kvCache driver.KVCacheProvider) {
	_initImageOnce.Do(func() {
		_image = &image{
			kvCache: kvCache,
		}
	})
}

func Get() Image {
	return _image
}
