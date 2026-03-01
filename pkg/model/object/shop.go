package object

import (
	"time"

	"gorm.io/gorm"
)

type Shop struct {
	ID           string    `gorm:"size:128;primarykey"`
	Name         string    `gorm:"size:100;not null"`
	CurrencyCode string    `gorm:"size:10;not null"`
	Status       string    `gorm:"size:20;not null"`
	CreatedAt    time.Time `gorm:"index"`
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
