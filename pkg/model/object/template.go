package object

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Template struct {
	ID            string         `gorm:"size:40;primarykey"`
	ShopID        string         `gorm:"size:128;not null"`
	Name          string         `gorm:"size:100"`
	Data          datatypes.JSON `gorm:"type:jsonb;not null"`
	PublishedData datatypes.JSON `gorm:"type:jsonb"`
	PageType      string         `gorm:"type:varchar(50);not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
