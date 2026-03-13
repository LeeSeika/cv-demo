package account

import (
	"context"
	"errors"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"gorm.io/gorm"
)

func (a *account) GetAccountByID(ctx context.Context, id string) (*object.Account, error) {
	var obj object.Account
	err := a.db.WithContext(ctx).First(&obj, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBizError(errs.ErrResourceNotFound, "account not found")
		}
		return nil, err
	}
	return &obj, nil
}
