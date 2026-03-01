package object

import (
	"time"

	"gorm.io/gorm"
)

type PaymentProfile struct {
	ID        string         `gorm:"size:128;primarykey"`
	Name      string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"default:now()"`
	UpdatedAt time.Time      `gorm:"default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft-delete Time
	// Associations
	PaymentRuleSets []PaymentRuleSet `gorm:"foreignKey:PaymentProfileID;constraint:OnDelete:CASCADE"`
}
