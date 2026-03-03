package dto

import "github.com/leeseika/cv-demo/pkg/model/object"

type ImagePreuploadReq struct {
	AltText     string `json:"alt_text"`
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type ImageResponse struct {
	ID          string `json:"id"`
	ShopID      string `json:"shop_id"`
	AltText     string `json:"alt_text"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

func BuildImageResponse(obj *object.Image, url string) *ImageResponse {
	return &ImageResponse{
		ID:          obj.ID,
		ShopID:      obj.ShopID,
		AltText:     obj.AltText,
		ContentType: obj.ContentType,
		URL:         url,
	}
}
