package object

import (
	"time"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/utils/datatype"
)

type ProductVariant struct {
	ID              string `gorm:"primarykey;autoIncrement"`
	Price           int    `gorm:"default:0;not null"`
	SKU             string `gorm:"size:100"`
	Title           string `gorm:"size:255;not null"`
	SelectedOptions datatype.JSONSlice[*jsonmodel.SelectedOption]
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// reference
	ProductID string  `gorm:"size:128;index;not null"`
	Product   Product `gorm:"foreignkey:ProductID"`
}
