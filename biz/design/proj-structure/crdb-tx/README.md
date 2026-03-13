# 重构事务抛出 error 的方式

### 1. CockroachDB 处理锁冲突的方式

CockroachDB 的处理方式具体见[官方文档](https://www.cockroachlabs.com/docs/stable/demo-serializable)

简单地说，CockroachDB 提供了一个`crdbgorm`库，用于检测 gorm 抛出的事务 error 是否是 `SQLSTATE 40001` 错误。如果是，则在客户端侧做事务重试。`crdbgorm` 源代码片段如下：

```go
// ExecuteInTx runs fn inside tx. This method is primarily intended for internal
// use. See other packages for higher-level, framework-specific ExecuteTx()
// functions.
//
// *WARNING*: It is assumed that no statements have been executed on the
// supplied Tx. ExecuteInTx will only retry statements that are performed within
// the supplied closure (fn). Any statements performed on the tx before
// ExecuteInTx is invoked will *not* be re-run if the transaction needs to be
// retried.
//
// fn is subject to the same restrictions as the fn passed to ExecuteTx.
func ExecuteInTx(ctx context.Context, tx Tx, fn func() error) (err error) {

	// establish the retry policy
	retryPolicy := getRetryPolicy(ctx)
	// set up the retry policy state
	retryFunc := retryPolicy.NewRetry()
	/*
	  for 循环，不退出循环就继续重试事务
	*/
	for {
		/*
		  执行 fn() 业务闭包
		*/
		err = fn()
		if err == nil {
		/*
		  省略其他逻辑
		*/
		}

		/*
		  根据业务闭包返回的错误判断是否可以重试
		*/
		// We got an error; let's see if it's a retryable one and, if so, restart.
		if !errIsRetryable(err) {
			if releaseFailed {
				err = newAmbiguousCommitError(err)
			}
			return err
		}

		/*
		  延迟 delay 时间后继续 for 循环，实现事务重试
		*/
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
```

我们能看到，关键是 `errIsRetryable` 函数判断 error 是否能做重试，函数逻辑如下：

```go
func errIsRetryable(err error) bool {
	// We look for either:
	//  - the standard PG errcode SerializationFailureError:40001 or
	//  - the Cockroach extension errcode RetriableError:CR000. This extension
	//    has been removed server-side, but support for it has been left here for
	//    now to maintain backwards compatibility.
	code := errCode(err)
	return code == "CR000" || code == "40001"
}

func errCode(err error) string {
	var sqlErr errWithSQLState
	if errors.As(err, &sqlErr) {
		return sqlErr.SQLState()
	}

	return ""
}

// errWithSQLState is implemented by pgx (pgconn.PgError) and lib/pq
type errWithSQLState interface {
	SQLState() string
}
```

CockroachDB 使用的 `sql driver` 是兼容 `Postgres` 协议的 `"gorm.io/driver/postgres"`，所以 gorm 返回 `PgError` 一定实现了 `errWithSQLState` 接口：

```go
// PgError represents an error reported by the PostgreSQL server. See
// http://www.postgresql.org/docs/11/static/protocol-error-fields.html for
// detailed field description.
type PgError struct {
	Severity         string
	Code             string
	Message          string
	Detail           string
	Hint             string
	Position         int32
	InternalPosition int32
	InternalQuery    string
	Where            string
	SchemaName       string
	TableName        string
	ColumnName       string
	DataTypeName     string
	ConstraintName   string
	File             string
	Line             int32
	Routine          string
}

func (pe *PgError) Error() string {
	return pe.Severity + ": " + pe.Message + " (SQLSTATE " + pe.Code + ")"
}

// SQLState returns the SQLState of the error.
func (pe *PgError) SQLState() string {
	return pe.Code
}
```

我们需要做的就是保证 DAO 层返回原样的、未经封装的 gorm error，最终事务闭包将 gorm error 返回给 `crdbgorm` 库，`crdbgorm` 就能根据 error 判断是否需要执行重试

### 2. 业务抛出事务 error 的设计

在分层设计中，数据库事务是在 Service 层开启的，数据库操作是在 DAO 层进行的。我们会给 Service 层所有的 error 都包装为 `BizError` 的形式，给 error 赋予业务含义。

要想包装后的 `BizError` 也能被 `crdbgorm` 库正确地判断是否需要重试，我们要做到如下两点：

- DAO 层原样返回 gorm 抛出的 error
- `BizError` 结构体要实现 `errWithSQLState` 接口

为了实现以上两点，我们选择用**组合**的方式让 DAO 层返回的 error 内嵌到 `BizError` 结构体中，这样 `BizError` 也实现了 `errWithSQLState` 接口。`BizError` 代码如下：

```go
// pkg/utils/errs/biz.go
package errs

type BizError struct {
	code    BizCode
	message string
	err     error
}

func WrapBizError(code BizCode, message string, err error) *BizError {
	return &BizError{
		code:    code,
		message: message,
		err:     err,
	}
}
```

业务代码例子如下：

```go
// biz/design/proj-structure/crdb-tx/service/account/assign_shop.go
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
		// 检查 account 是否存在
		exists, err := a.accountDAO.CheckExists(ctx, accountID)
		if err != nil {
			// WrapBizError 把 CheckExists 返回的 DAO error 内嵌到 BizError 中
			return errs.WrapBizError(errs.ErrInternalServer, "failed to check account", err)
		}
		if !exists {
			return errs.NewBizError(errs.ErrResourceNotFound, "account not found")
		}

		// 指派 shop 给 account
		rowsAffected, err := a.shopDAO.UpdateAssignAccount(ctx, shopID, accountID)
		if err != nil {
			// WrapBizError 把 UpdateAssignAccount 返回的 DAO error 内嵌到 BizError 中
			return errs.WrapBizError(errs.ErrInternalServer, "failed to update shop", err)
		}
		if rowsAffected == 0 {
			return errs.NewBizError(errs.ErrResourceNotFound, "shop not found")
		}

		return nil
	})

	return err
}
```

按照这样的设计，`crdbgorm` 库接收到业务闭包返回的 `errs.WrapBizError()` 生成的 `BizError` 时，`BizError` 通过**组合**的方式内嵌了 gorm error（也就是 `PgError`），当 `crdbgorm` 调用 `BizError` 的 `SQLState()` 方法时，能够正确地拿到 gorm error 返回的 SQLState。
