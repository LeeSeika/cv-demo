package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) GetAccountByID(ctx context.Context, id string) (*object.Account, error) {
	var obj object.Account
	err := a.db.WithContext(ctx).First(&obj, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
