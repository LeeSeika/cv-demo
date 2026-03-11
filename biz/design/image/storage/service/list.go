package service

import (
	"context"
	"errors"
	"strings"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (i *image) List(ctx context.Context, shopID string) ([]*object.Image, error) {
	shopID = strings.TrimSpace(shopID)
	if len(shopID) == 0 {
		return nil, errors.New("shop id is required")
	}
	if i.storageProvider == nil {
		return nil, errors.New("storage provider is not initialized")
	}
	if i.db == nil {
		return nil, errors.New("db is not initialized")
	}

	var images []*object.Image
	err := i.db.WithContext(ctx).
		Model(&object.Image{}).
		Where("shop_id = ?", shopID).
		Where("bucket = ?", i.storageProvider.Bucket()).
		Where("is_uploaded = ?", true).
		Order("created_at DESC").
		Find(&images).Error
	if err != nil {
		return nil, err
	}

	return images, nil
}
