package object

import "time"

type Image struct {
	ID          string `gorm:"size:128;primaryKey"`
	ShopID      string `gorm:"size:128;index"`
	AltText     string `gorm:"size:256"`
	ContentType string `gorm:"size:256"`
	Bucket      string `gorm:"size:256;index"`
	FileKey     string `gorm:"size:512;index"`
	IsUploaded  bool   `gorm:"not null;default:false;index"`
	UploadedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
