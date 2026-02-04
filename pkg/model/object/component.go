package object

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Component struct {
	Version       string         `gorm:"size:60;primarykey"`
	Name          string         `gorm:"size:100;primarykey"`
	Liquid        string         `gorm:"type:text;not null"`
	Schema        datatypes.JSON `gorm:"type:jsonb;not null"`
	ScriptBucket  string         `gorm:"size:40"`
	ScriptFileKey string         `gorm:"size:400"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
