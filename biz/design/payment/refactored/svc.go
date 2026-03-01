package refactored

import (
	"context"

	"gorm.io/gorm"
)

type (
	PaymentRule interface {
		Update(ctx context.Context, ruleID string)
	}

	paymentRule struct {
		db *gorm.DB
	}
)
