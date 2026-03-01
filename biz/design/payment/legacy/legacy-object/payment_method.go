package legacyobject

import (
	"time"

	"github.com/leeseika/cv-demo/pkg/constants"
	"gorm.io/datatypes"
)

type PaymentMethod struct {
	ID          string              `gorm:"size:128;primarykey"`
	Provider    constants.Provider  `gorm:"size:128;not null;index"`
	PayMethod   constants.PayMethod `gorm:"type:varchar(50);not null;index"`
	Credentials datatypes.JSON      `gorm:"type:jsonb;not null"`
	CreatedAt   time.Time           `gorm:"default:now()"`
	UpdatedAt   time.Time           `gorm:"default:now()"`
	Position    int                 `gorm:"not null;default:1"`
	ZoneID      string              `gorm:"size:128;not null;index"`
}
