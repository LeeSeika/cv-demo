package object

import (
	"time"

	"github.com/leeseika/cv-demo/pkg/constants"
	"gorm.io/datatypes"
)

type PaymentRule struct {
	ID         string             `gorm:"size:128;primarykey"`
	Name       string             `gorm:"type:varchar(255);not null"`
	Type       constants.RuleType `gorm:"type:varchar(50);not null;uniqueIndex:idx_rule_set_id_and_type,priority:2,where:(type = 'default_routing')"`
	RuleSetID  string             `gorm:"size:128;not null;uniqueIndex:idx_rule_set_id_and_type,priority:1,where:(type = 'default_routing')"`
	Selection  datatypes.JSON     `gorm:"type:jsonb;not null"`
	Statements datatypes.JSON     `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time          `gorm:"default:now()"`
	UpdatedAt  time.Time          `gorm:"default:now()"`
	// Associations
	PaymentRuleSet PaymentRuleSet `gorm:"foreignKey:RuleSetID;constraint:OnDelete:CASCADE"`
}
