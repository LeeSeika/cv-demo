package legacy

import (
	"context"

	legacyobject "github.com/leeseika/cv-demo/biz/design/payment/legacy/legacy-object"
	"gorm.io/gorm"
)

func (s *paymentMethod) UpdatePaymentMethodSort(ctx context.Context, zoneID string, methodID string, newPosition int) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var paymentMethod legacyobject.PaymentMethod
		err := tx.Model(&legacyobject.PaymentMethod{}).
			Where("payment_methods.id = ?", methodID).
			Where("payment_methods.zone_id = ?", zoneID).
			First(&paymentMethod).Error
		if err != nil {
			return err
		}

		if newPosition < 1 {
			newPosition = 1
		}

		originalPosition := paymentMethod.Position

		var maxPosition int
		err = tx.Model(&legacyobject.PaymentMethod{}).
			Where("zone_id = ?", zoneID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxPosition).Error
		if err != nil {
			return err
		}

		if newPosition > maxPosition {
			newPosition = maxPosition
		}

		if originalPosition == newPosition {
			return nil
		}

		if originalPosition < newPosition {
			err = tx.Model(&legacyobject.PaymentMethod{}).
				Where("zone_id = ?", zoneID).
				Where("position > ?", originalPosition).
				Where("position <= ?", newPosition).
				UpdateColumn("position", gorm.Expr("position - 1")).Error
			if err != nil {
				return err
			}
		} else {
			err = tx.Model(&legacyobject.PaymentMethod{}).
				Where("zone_id = ?", zoneID).
				Where("position >= ?", newPosition).
				Where("position < ?", originalPosition).
				UpdateColumn("position", gorm.Expr("position + 1")).Error
			if err != nil {
				return err
			}
		}

		err = tx.Model(&legacyobject.PaymentMethod{}).
			Where("id = ?", paymentMethod.ID).
			UpdateColumn("position", newPosition).Error
		if err != nil {
			return err
		}

		return nil
	})
	return err
}
