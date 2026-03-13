package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) CreateAccount(ctx context.Context, req *dto.CreateAccountReq) (string, error) {
	id := uuid.New().String()
	obj := &object.Account{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	if err := a.dao.CreateAccount(ctx, obj); err != nil {
		return "", err
	}
	return id, nil
}
