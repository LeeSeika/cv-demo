package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
)

func (a *account) UpdateAccount(ctx context.Context, id string, req *dto.UpdateAccountReq) error {
	obj := &object.Account{
		Name:        req.Name,
		AvatarURL:   req.AvatarURL,
		Description: req.Description,
	}
	rowsAffected, err := a.dao.UpdateAccount(ctx, id, obj)
	if err != nil {
		return errs.NewBizError(errs.ErrInternalServer, "failed to update account")
	}
	// 在 Service 层给 DAO 返回的结果赋予业务含义
	if rowsAffected == 0 {
		return errs.NewBizError(errs.ErrResourceNotFound, "account not found")
	}
	return nil
}
