package legacy

import (
	"context"

	"gorm.io/gorm"
)

type (
	PaymentMethod interface {
		UpdatePaymentMethodSort(ctx context.Context, paymentProfileID string, zoneID string, methodID string, newPosition int) error
	}

	paymentMethod struct {
		db *gorm.DB
	}
)
