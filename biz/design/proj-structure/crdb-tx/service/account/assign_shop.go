package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"github.com/leeseika/cv-demo/pkg/utils/transaction"
)

func (a *account) AssignShop(ctx context.Context, accountID string, shopID string) error {
	tx := transaction.NewHandler()

	// 开启事务
	err := tx.Run(ctx, nil, func(ctx context.Context) error {
		// check account exists
		exists, err := a.accountDAO.CheckExists(ctx, accountID)
		if err != nil {
			return errs.WrapBizError(errs.ErrInternalServer, "failed to check account", err)
		}
		if !exists {
			return errs.NewBizError(errs.ErrResourceNotFound, "account not found")
		}

		// update shop assigned account
		rowsAffected, err := a.shopDAO.UpdateAssignAccount(ctx, shopID, accountID)
		if err != nil {
			return errs.WrapBizError(errs.ErrInternalServer, "failed to update shop", err)
		}
		if rowsAffected == 0 {
			return errs.NewBizError(errs.ErrResourceNotFound, "shop not found")
		}

		return nil
	})

	return err
}
