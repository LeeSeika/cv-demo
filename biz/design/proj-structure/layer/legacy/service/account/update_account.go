package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) UpdateAccount(ctx context.Context, id string, req *dto.UpdateAccountReq) error {
	obj := &object.Account{
		Name:        req.Name,
		AvatarURL:   req.AvatarURL,
		Description: req.Description,
	}
	return a.dao.UpdateAccount(ctx, id, obj)
}
