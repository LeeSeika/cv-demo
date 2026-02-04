package object

import "time"

type Image struct {
	ID          string    `gorm:"size:128;primaryKey"`
	AltText     string    `gorm:"size:256"`
	ContentType string    `gorm:"size:256"`
	Bucket      string    `gorm:"size:256;index"`
	OriginalSrc string    `gorm:"size:512;index"`
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}
