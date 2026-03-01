package object

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Template struct {
	ID        string         `gorm:"size:40;primarykey"`
	Name      string         `gorm:"size:100"`
	Data      datatypes.JSON `gorm:"type:jsonb;not null"`
	Version   int            `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
