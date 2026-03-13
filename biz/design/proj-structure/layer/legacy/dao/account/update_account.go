package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
)

func (a *account) UpdateAccount(ctx context.Context, id string, obj *object.Account) error {
	result := a.db.WithContext(ctx).Model(&object.Account{}).Where("id = ?", id).Updates(obj)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.NewBizError(errs.ErrResourceNotFound, "account not found")
	}
	return nil
}
