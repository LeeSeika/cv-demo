package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/transaction"
)

func (a *account) CheckExists(ctx context.Context, id string) (bool, error) {
	db := transaction.GetExecDB(ctx, a.db)

	var count int64
	if err := db.Model(&object.Account{}).
		Where("id = ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
