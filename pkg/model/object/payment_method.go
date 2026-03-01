package object

import (
	"time"

	"github.com/leeseika/cv-demo/pkg/constants"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentMethod struct {
	ID               string              `gorm:"size:128;primarykey"`
	PaymentGatewayID string              `gorm:"size:128;not null;index" `
	PayMethod        constants.PayMethod `gorm:"type:varchar(50);not null;index"`
	Credentials      datatypes.JSON      `gorm:"type:jsonb;not null"`
	CreatedAt        time.Time           `gorm:"default:now()"`
	UpdatedAt        time.Time           `gorm:"default:now()"`
	DeletedAt        gorm.DeletedAt      `gorm:"index"` // Soft-delete Time

	PaymentGateway PaymentGateway `gorm:"foreignKey:PaymentGatewayID;references:ID"`
}
