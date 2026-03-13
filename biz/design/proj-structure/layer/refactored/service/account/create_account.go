package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"gorm.io/gorm"
)

func (a *account) CreateAccount(ctx context.Context, req *dto.CreateAccountReq) (string, error) {
	id := uuid.New().String()
	obj := &object.Account{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	err := a.dao.CreateAccount(ctx, obj)
	if err != nil {
		if errs.IsDBError(err, gorm.ErrDuplicatedKey) {
			return "", errs.NewBizError(errs.ErrResourceExists, "account already exists")
		}
		return "", errs.NewBizError(errs.ErrInternalServer, "failed to create account")
	}
	return id, nil
}
