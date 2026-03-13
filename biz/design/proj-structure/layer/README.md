# 重构项目结构

## 1. 重构原则

**各层职责**

- API 层：解析请求参数，组装 API 响应
- Service 层：调用 DAO、driver 等下层接口，执行业务操作；封装下层抛出的 error，给 error 赋予业务意义
- 数据访问层（DAO）：操作数据，原样返回操作结果

**重构原则**

- 上层不应该关心下层接口的具体实现
- 下层不应该关心上层业务逻辑

## 1. 重构返回值

### 1.1. 原项目的设计

在原来的项目结构中，Service 层的返回值结构体通常就是 API 层的响应结构体，例子如下：

**Service 层**

```go
// biz/design/proj-structure/layer/legacy/service/account/get_account_by_id.go
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
```

**API 层**

```go
// biz/design/proj-structure/layer/legacy/api/account/get_by_id.go
package account

import (
	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/proj-structure/layer/legacy/service/account"
	"github.com/leeseika/cv-demo/pkg/utils/api"
)

func GetByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// 直接响应 Service 层返回的结构体
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", info))
}
```

**model 层（DTO）**

```go
// pkg/model/dto/account.go
package dto

type AccountInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
```

**model 层（database object）**

```go
// pkg/model/object/account.go
package object

import "time"

type Account struct {
	ID          string `gorm:"size:128;primarykey"`
	Name        string `gorm:"size:128;not null"`
	Email       string `gorm:"size:320;uniqueIndex"`
	Password    string `gorm:"size:256;not null"`
	TOTPSecret  string `gorm:"size:512;not null"`
	AvatarURL   string `gorm:"size:512;not null"`
	Description string `gorm:"size:512;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

在上述代码中，组装 DTO 的工作下放到了 Service 层。假设现在有新的 `GetCompleteAccountInfo` API，想要获取以下结构的 Account 信息：

```go
type CompleteAccountInfo struct {
	ID          string
	Name        string
	Email       string
	AvatarURL   string
	Description string
}
```

新 API 逻辑跟原 API 完全一致，都是获取 AccountInfo，只是新 API 的 HTTP Response 需要比原 API 多`email`和`description`两个字段。

如果我们坚持 DTO 在 Service 层组装的设计，那么原有的 `GetAccountByID` 方法就不能复用，为了新增两个字段只能重写一个全新的 `GetCompleteAccountByID` 方法，而且方法的业务逻辑基本上是一致的：

```go
// 新创建的 GetCompleteAccountByID Service 方法，返回完整的 AccountInfo
func (a *account) GetCompleteAccountByID(ctx context.Context, id string) (*dto.CompleteAccountInfo, error) {
	obj, err := a.dao.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.CompleteAccountInfo{
		ID:          obj.ID,
		Name:        obj.Name,
		Email:       obj.Email,
		AvatarURL:   obj.AvatarURL,
		Description: obj.Description,
	}, nil
}
```

或者，就算我们不新增 `GetCompleteAccountByID` 方法，选择继续复用原 `GetAccountByID` 方法，通过修改 `GetAccountByID` 方法的返回值，让原 `GetAccountByID` 返回完整的 `*dto.CompleteAccountInfo` 信息，这样也会导致 API 层职责不一致：

**修改后的 GetAccountByID Service 方法**

```go
// 修改原 GetAccountByID Service 方法，返回完整的 CompleteAccountInfo 信息
func (a *account) GetAccountByID(ctx context.Context, id string) (*dto.CompleteAccountInfo, error) {
	obj, err := a.dao.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.CompleteAccountInfo{
		ID:          obj.ID,
		Name:        obj.Name,
		Email:       obj.Email,
		AvatarURL:   obj.AvatarURL,
		Description: obj.Description,
	}, nil
}
```

**原 API（返回值只需要 id、email、avatar_url 三个字段）**

```go
// 修改原 GetByID API，从 CompleteAccountInfo 中抽取需要的字段
func GetByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// GetByID API 只需要以下三个字段，在 API 层重新组装了 HTTP Response
	data := gin.H{
		"id":         info.ID,
		"name":       info.Name,
		"avatar_url": info.AvatarURL,
	}
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", data))
}
```

**新 API（返回值需要 id、email、avatar_url、email、description 五个字段）**

```go
func GetCompleteByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// 直接返回完整的账户信息，API 层没有重新组装 HTTP Response
	c.JSON(200, api.SuccessResponse("Complete account retrieved successfully", info))
}
```

可以看到，即使共享了同一个 Service 方法，还是会导致 API 层职责不一致，一个 API 重新组装了 HTTP Response DTO，另一个 API 则没有组装。

因为这种设计把 API 响应的结构体绑定在了 Service 层的返回值上，只要 Service 层的返回值发生变动，一定会影响到 API 的响应。API 层和 Service 层的职责不清晰，导致两者耦合在了一起。

### 1.2. 重构

为了解决上述问题，我们允许 Service 层在需要的时候定义自身的返回值结构体。当 Service 返回值结构体跟 DAO 返回的结构体完全一致时，也可以直接使用 DAO 返回的结构体，避免重复定义。
重构后，`GetAccountByID` 和 `GetCompleteAccountByID` 两个 API 的逻辑如下

**Service 层**

```go
package account

import (
	"context"
	"time"

	"github.com/leeseika/cv-demo/pkg/utils/errs"
	"gorm.io/gorm"
)

// AccountInfo 自定义 Service 返回值，去除敏感字段 password、totp_secret
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
```

**API 层**

```go
// GetByID 返回 id、name、avatar_url
func GetByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountInfoByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// API 层组装 DTO，提取需要返回的字段
	responseAccountInfo := &dto.AccountInfo{
		ID:        info.ID,
		Name:      info.Name,
		AvatarURL: info.AvatarURL,
	}
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", responseAccountInfo))
}

// GetCompleteByID 返回 id、name、avatar_url、email、description
func GetCompleteByID(c *gin.Context) {
	id := c.Param("id")
	info, err := accountsvc.Get().GetAccountInfoByID(c, id)
	if err != nil {
		statusCode, resp := api.BizErrorResponse(err)
		c.JSON(statusCode, resp)
		return
	}
	// API 层组装 DTO，提取需要返回的字段
	responseAccountInfo := &dto.CompleteAccountInfo{
		ID:          info.ID,
		Name:        info.Name,
		AvatarURL:   info.AvatarURL,
		Email:       info.Email,
		Description: info.Description,
	}
	c.JSON(200, api.SuccessResponse("Account retrieved successfully", responseAccountInfo))
}
```

重构后，各层职责明确、清晰，降低了耦合度。

## 2. 重构 error

### 2.1. 原项目的设计

在原项目中，通常会在 DAO 层给 error 赋予业务含义。假设在更新用户的场景下，请求参数携带的用户 id 在数据库中不存在，原项目的处理方式如下：

**DAO 层**

```go
// biz/design/proj-structure/layer/legacy/dao/account/update_account.go
package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/errs"
)

func (a *account) UpdateAccount(ctx context.Context, id string, obj *object.Account) error {
	result := a.db.WithContext(ctx).Model(&object.Account{}).Where("id = ?", id).Updates(obj)
	if result.Error != nil {
		return result.Error
	}
	// 在 DAO 层封装业务 error
	if result.RowsAffected == 0 {
		return errs.NewBizError(errs.ErrResourceNotFound, "account not found")
	}
	return nil
}
```

**Service 层**

```go
// biz/design/proj-structure/layer/legacy/service/account/update_account.go
ppackage account

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
```

在这种设计下，**数据访问层**擅自给数据访问结果赋予了业务含义。在这个业务例子中，“访问数据库”这个动作并没有出错，但是 DAO 层根据数据库返回的 `rows affected` 结果封装了业务错误。导致 DAO 层完成了本属于 Service 层的工作，两层的职责出现了耦合。

**tips**：这种设计还会导致一个问题，因为 DAO 层重新封装了业务 error，导致 crdbgorm 库无法根据事务闭包返回的 error 判断是否需要重试，具体内容见 [crdb-tx](/biz/design/proj-structure/crdb-tx/)

### 2.2. 重构

我们让 DAO 层专注于**数据访问**职责，当 DAO 层出现错误时，一定是**数据访问**这个动作出现了错误，它跟上层业务逻辑无关。重构后的例子如下：

**DAO 层**

```go
// biz/design/proj-structure/layer/refactored/dao/account/update_account.go
package account

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *account) UpdateAccount(ctx context.Context, id string, obj *object.Account) (int64, error) {
	result := a.db.WithContext(ctx).Model(&object.Account{}).Where("id = ?", id).Updates(obj)
	if result.Error != nil {
		return 0, result.Error
	}
	// 原样返回数据访问层结果
	return result.RowsAffected, nil
}
```

**Service 层**

```go
// biz/design/proj-structure/layer/refactored/service/account/update_account.go
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
```
