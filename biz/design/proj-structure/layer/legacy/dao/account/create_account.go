package account

import (
	"context"
	"errors"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"gorm.io/gorm"
)

func (a *account) CreateAccount(ctx context.Context, obj *object.Account) error {
	err := a.db.WithContext(ctx).Create(obj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.NewBizError(errs.ErrResourceExists, "account already existed")
		}
		return err
	}
	return nil
}
