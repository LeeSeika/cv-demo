package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (i *image) ConfirmUpload(ctx context.Context, shopID string, imageID string) error {
	shopID = strings.TrimSpace(shopID)
	imageID = strings.TrimSpace(imageID)

	if len(shopID) == 0 {
		return errors.New("shop id is required")
	}
	if len(imageID) == 0 {
		return errors.New("image id is required")
	}
	if i.db == nil {
		return errors.New("db is not initialized")
	}
	if i.storageProvider == nil {
		return errors.New("storage provider is not initialized")
	}

	var image object.Image
	err := i.db.WithContext(ctx).
		Where("id = ?", imageID).
		Where("shop_id = ?", shopID).
		First(&image).Error
	if err != nil {
		return errors.New("image not found")
	}

	exists, err := i.storageProvider.ObjectExists(ctx, image.FileKey)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("image object not found in oss")
	}

	now := time.Now()
	result := i.db.WithContext(ctx).
		Model(&object.Image{}).
		Where("id = ?", image.ID).
		Where("shop_id = ?", image.ShopID).
		Updates(map[string]any{
			"is_uploaded": true,
			"uploaded_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("image not found")
	}

	return nil
}