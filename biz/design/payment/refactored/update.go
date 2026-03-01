package refactored

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/datatypes"
)

type UpdatePaymentRuleInput struct {
	Name       string
	Selection  json.RawMessage
	Statements json.RawMessage
}

func (s *paymentRule) Update(ctx context.Context, ruleID string, input UpdatePaymentRuleInput) error {
	updatedRule := &object.PaymentRule{
		Name:       input.Name,
		Selection:  datatypes.JSON(input.Selection),
		Statements: datatypes.JSON(input.Statements),
	}
	err := s.db.WithContext(ctx).Model(&object.PaymentRule{}).Where("id = ?", ruleID).Updates(updatedRule).Error
	if err != nil {
		return err
	}
	return nil
}
