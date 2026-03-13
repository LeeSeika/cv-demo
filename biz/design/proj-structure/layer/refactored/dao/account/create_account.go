package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) CreateAccount(ctx context.Context, obj *object.Account) error {
	err := a.db.WithContext(ctx).Create(obj).Error
	if err != nil {
		return err
	}
	return nil
}
