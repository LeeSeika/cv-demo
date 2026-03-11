package legacyobject

import "gorm.io/datatypes"

type RoutingZone struct {
	ID                string         `gorm:"size:128;primarykey"`
	Name              string         `gorm:"type:varchar(255);not null"`
	SelectedCountries datatypes.JSON `gorm:"type:jsonb;not null"`
	PaymentProfileID  string         `gorm:"size:128;not null;index"`
	// Associations
	PaymentMethods []PaymentMethod `gorm:"foreignKey:ZoneID;constraint:OnDelete:CASCADE"`
}
