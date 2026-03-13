package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) UpdateAccount(ctx context.Context, id string, obj *object.Account) (int64, error) {
	result := a.db.WithContext(ctx).Model(&object.Account{}).Where("id = ?", id).Updates(obj)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
