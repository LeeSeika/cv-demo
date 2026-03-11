package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

const defaultPreuploadExpireDuration = 15 * time.Minute

func (i *image) Preupload(ctx context.Context, shopID string, req *dto.ImagePreuploadReq) (*dto.ImagePreuploadResponse, error) {
	shopID = strings.TrimSpace(shopID)
	filename := strings.TrimSpace(req.Filename)
	contentType := strings.TrimSpace(req.ContentType)

	if len(shopID) == 0 {
		return nil, errors.New("shop id is required")
	}
	if len(filename) == 0 {
		return nil, errors.New("filename is required")
	}
	if len(contentType) == 0 {
		return nil, errors.New("content type is required")
	}

	imageID := uuid.New().String()
	fileKey := i.storageURLBuilder.BuildFileKey(shopID, imageID, filename)
	uploadURL, err := i.storageProvider.GeneratePresignedUploadURL(ctx, fileKey, contentType, defaultPreuploadExpireDuration)
	if err != nil {
		return nil, err
	}

	bucket := i.storageProvider.Bucket()
	image := &object.Image{
		ID:          imageID,
		ShopID:      shopID,
		AltText:     req.AltText,
		ContentType: contentType,
		Bucket:      bucket,
		FileKey:     fileKey,
	}
	if createErr := i.db.WithContext(ctx).Create(image).Error; createErr != nil {
		return nil, createErr
	}

	return &dto.ImagePreuploadResponse{
		ImageID:   imageID,
		UploadURL: uploadURL,
		Method:    "PUT",
	}, nil
}
