package batchupsert

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductVariantService struct {
	db *gorm.DB
}

func NewProductVariantService(db *gorm.DB) *ProductVariantService {
	return &ProductVariantService{db: db}
}

func (s *ProductVariantService) UpsertVariants_OnConflict(ctx context.Context, variants []object.ProductVariant) error {
	db := s.db.WithContext(ctx)

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "price", "updated_at"}),
	}).Create(&variants).Error
}

func (s *ProductVariantService) UpsertVariants_ReplaceInto(ctx context.Context, variants []object.ProductVariant) error {
	db := s.db.WithContext(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		ids := make([]string, 0, len(variants))
		for _, variant := range variants {
			ids = append(ids, variant.ID)
		}
		if err := tx.Where("id IN ?", ids).Delete(&object.ProductVariant{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&variants).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}
