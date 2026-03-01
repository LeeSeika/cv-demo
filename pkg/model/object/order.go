package object

import "time"

type Order struct {
	ID          string    `gorm:"size:128;primarykey"`
	ShopID      string    `gorm:"size:128;index;not null"`
	OrderNo     string    `gorm:"size:64;index;not null"`
	Status      string    `gorm:"size:32;not null"`
	TotalAmount int64     `gorm:"not null"`
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}
