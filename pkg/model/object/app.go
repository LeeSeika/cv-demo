package object

import (
	"time"

	"gorm.io/gorm"
)

type App struct {
	ID              string `gorm:"size:128;primarykey"`
	Name            string `gorm:"size:256;index;not null"`
	Secret          string `gorm:"size:64;not null"`
	InstallationURL string `gorm:"size:512"`
	RedirectURL     string `gorm:"size:512"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
