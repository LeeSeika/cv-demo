package object

import "time"

type ShopAccount struct {
	AccountID string `gorm:"size:128;primarykey"`
	ShopID    string `gorm:"size:128;primarykey"`
	Role      string `gorm:"size:16;check:role='OWNER' OR role='ADMIN' OR role='STAFF' "`
	CreatedAt time.Time
	UpdatedAt time.Time
	Account   Account `gorm:"foreignkey:AccountID;constraint:OnDelete:CASCADE"`
	Shop      Shop    `gorm:"foreignkey:ShopID;constraint:OnDelete:CASCADE"`
}
