package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) GetAccountInfo(ctx context.Context, accountID string) (*object.Account, error) {
	var account object.Account
	err := a.db.WithContext(ctx).First(&account, "id = ?", accountID).Error
	if err != nil {
		return nil, err
	}

	return &account, nil
}
