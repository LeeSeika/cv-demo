package service

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _image Image
var _imageOnce sync.Once

type (
	Image interface {
		Preupload(ctx context.Context, shopID string, req *dto.ImagePreuploadReq) (string, error)
		List(ctx context.Context, shopID string) ([]*object.Image, error)
	}

	image struct {
		db                *gorm.DB
		storageProvider   driver.StorageProvider
		storageURLBuilder driver.StorageURLBuilder
	}
)

func Init() {
	_imageOnce.Do(func() {
		_image = &image{
			db:                driver.GetDB(),
			storageProvider:   driver.GetStorageProvider(),
			storageURLBuilder: driver.GetStorageURLBuilder(),
		}
	})
}

func Get() Image {
	return _image
}
