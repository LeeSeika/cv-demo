package image

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (i *image) GetByIDs(ctx context.Context, imageIDs []string) ([]*object.Image, error) {
	db := i.db.WithContext(ctx)

	var images []*object.Image
	res := db.Model(object.Image{}).
		Where("id IN ?", imageIDs).
		Find(&images)
	if res.Error != nil {
		return nil, res.Error
	}
	return images, nil
}
