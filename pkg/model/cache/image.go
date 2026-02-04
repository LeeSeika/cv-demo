package cache

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

const storageDomain = "http://localhost:4001"

type Image struct {
	ID      string `json:"id"`
	AltText string `json:"alt_text"`
	Src     string `json:"src"`
}

func ImageFromObject(obj *object.Image) *Image {
	return &Image{
		ID:      obj.ID,
		AltText: obj.AltText,
		Src:     fmt.Sprintf("%s/%s/%s", storageDomain, obj.Bucket, obj.OriginalSrc),
	}
}
