package legacyobject

import (
	"time"
)

type PaymentProfile struct {
	ID        string    `gorm:"size:128;primarykey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"default:now()"`
	// Associations
	Zones []RoutingZone `gorm:"foreignKey:PaymentProfileID;constraint:OnDelete:CASCADE"`
}
