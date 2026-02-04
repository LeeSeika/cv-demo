package object

type ProductImage struct {
	ProductID string  `gorm:"size:128;primaryKey"`
	ImageID   string  `gorm:"size:128;primaryKey"`
	Product   Product `gorm:"foreignkey:ProductID;constraint:OnDelete:CASCADE"`
	Image     Image   `gorm:"foreignkey:ImageID;constraint:OnDelete:CASCADE"`
}
