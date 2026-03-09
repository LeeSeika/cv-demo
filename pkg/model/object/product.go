package object

import (
	"database/sql"
	"time"

	"github.com/leeseika/cv-demo/pkg/datatype"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"gorm.io/gorm"
)

type Product struct {
	ID          string `gorm:"size:128;primarykey"`
	ShopID      string `gorm:"size:128"`
	Title       string `gorm:"size:128;index;force"`
	Description string `gorm:"size:10240"`
	Options     datatype.JSONSlice[*jsonmodel.ProductOption]
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	// references
	TemplateID      sql.NullString   `gorm:"size:128;index;constraint:OnUpdate:CASCADE;constraint:OnDelete:SET NULL"`
	Template        Template         `gorm:"foreignkey:TemplateID"`                            // The template associated with this product
	ProductVariants []ProductVariant `gorm:"foreignkey:ProductID;constraint:OnDelete:CASCADE"` // The product variants associated with this product
}
