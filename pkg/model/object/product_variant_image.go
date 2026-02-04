package object

type ProductVariantImage struct {
	ProductVariantID string         `gorm:"size:128;primaryKey"`
	ImageID          string         `gorm:"size:128;primaryKey"`
	ProductVariant   ProductVariant `gorm:"foreignkey:ProductVariantID;constraint:OnDelete:CASCADE"`
	Image            Image          `gorm:"foreignkey:ImageID;constraint:OnDelete:CASCADE"`
}
