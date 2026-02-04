package image

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _image Image
var _initImageOnce sync.Once

type (
	Image interface {
		GetByIDs(ctx context.Context, imageIDs []string) ([]*object.Image, error)
	}

	image struct {
		db *gorm.DB
	}
)

func Init(db *gorm.DB) {
	_initImageOnce.Do(func() {
		_image = &image{
			db: db,
		}
	})
}

func Get() Image {
	return _image
}
