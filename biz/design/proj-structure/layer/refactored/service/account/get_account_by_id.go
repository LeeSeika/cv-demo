package account

import (
	"context"
	"time"

	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"gorm.io/gorm"
)

type AccountInfo struct {
	ID          string
	Name        string
	Email       string
	AvatarURL   string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (a *account) GetAccountInfoByID(ctx context.Context, id string) (*AccountInfo, error) {
	obj, err := a.dao.GetAccountByID(ctx, id)
	if err != nil {
		if errs.IsDBError(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBizError(errs.ErrResourceNotFound, "account not found")
		}
		return nil, errs.NewBizError(errs.ErrInternalServer, "failed to get account")
	}
	// 定义对外可见字段
	return &AccountInfo{
		ID:          obj.ID,
		Name:        obj.Name,
		Email:       obj.Email,
		AvatarURL:   obj.AvatarURL,
		Description: obj.Description,
		CreatedAt:   obj.CreatedAt,
		UpdatedAt:   obj.UpdatedAt,
	}, nil
}
