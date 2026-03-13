package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (a *account) GetAccountByID(ctx context.Context, id string) (*dto.AccountInfo, error) {
	obj, err := a.dao.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.AccountInfo{
		ID:        obj.ID,
		Name:      obj.Name,
		AvatarURL: obj.AvatarURL,
	}, nil
}
