package object

import (
	"time"

	"gorm.io/datatypes"
)

type ShopApp struct {
	AppID        string `gorm:"size:128;primarykey;foreignkey:ID"`
	ShopID       string `gorm:"size:128;primarykey;foreignkey:ID"`
	Status       string
	AuthCode     string
	AccessToken  string
	RefreshToken string
	TokenInfo    datatypes.JSON
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
