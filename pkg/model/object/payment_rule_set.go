package object

import (
	"github.com/leeseika/cv-demo/pkg/constants"
	"gorm.io/datatypes"
)

type PaymentRuleSet struct {
	ID               string              `gorm:"size:128;primarykey"`
	Name             string              `gorm:"type:varchar(255);not null"`
	PaymentProfileID string              `gorm:"size:128;not null;uniqueIndex:idx_profile_id_and_pay_method,priority:1"`
	PayMethod        constants.PayMethod `gorm:"type:varchar(50);not null;uniqueIndex:idx_profile_id_and_pay_method,priority:2"`
	FlowGraph        datatypes.JSON      `gorm:"type:jsonb;not null"`
	// Associations
	PaymentRules []PaymentRule `gorm:"foreignKey:PaymentRuleSetID;references:ID"`
}
