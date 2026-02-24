package object

import "time"

type ShopApp struct {
	AppID     string `gorm:"size:128;primarykey;foreignkey:ID"`
	ShopID    string `gorm:"size:128;primarykey;foreignkey:ID"`
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
