package shop

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/transaction"
)

func (s *shop) UpdateAssignAccount(ctx context.Context, shopID string, accountID string) (int64, error) {
	db := transaction.GetExecDB(ctx, s.db)

	result := db.Model(&object.Shop{}).
		Where("id = ?", shopID).
		Update("assigned_account_id", accountID)

	return result.RowsAffected, result.Error
}
