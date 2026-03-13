package account

import (
	"context"
	"sync"

	accountdao "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/refactored/dao/account"
	"github.com/leeseika/cv-demo/pkg/model/dto"
)

var _account *account
var _initAccountOnce sync.Once

type (
	Account interface {
		CreateAccount(ctx context.Context, req *dto.CreateAccountReq) (string, error)
		GetAccountInfoByID(ctx context.Context, id string) (*AccountInfo, error)
		UpdateAccount(ctx context.Context, id string, req *dto.UpdateAccountReq) error
	}

	account struct {
		dao accountdao.Account
	}
)

func Init() {
	_initAccountOnce.Do(func() {
		_account = &account{
			dao: accountdao.Get(),
		}
	})
}

func Get() Account {
	return _account
}
