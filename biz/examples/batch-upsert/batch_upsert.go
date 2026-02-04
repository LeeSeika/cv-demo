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

func (s *ProductVariantService) UpsertVariants(ctx context.Context, variants []object.ProductVariant) error {
	db := s.db.WithContext(ctx)

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "price", "updated_at"}),
	}).Create(&variants).Error
}
