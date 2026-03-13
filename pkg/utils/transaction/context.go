package transaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type contextKey struct{}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, contextKey{}, tx)
}

func GetExecDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(contextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return defaultDB
}

// CreateSavePoint create tx save point
func CreateSavePoint(ctx context.Context, name string) error {
	tx := GetExecDB(ctx, nil) // get current tx
	if tx == nil {
		return fmt.Errorf("no active transaction")
	}

	// GORM save tx
	return tx.SavePoint(name).Error
}

// RollbackToSavePoint rollback to tx save point
func RollbackToSavePoint(ctx context.Context, name string) error {
	tx := GetExecDB(ctx, nil)
	if tx == nil {
		return fmt.Errorf("no active transaction")
	}

	// GORM rollback tx
	return tx.RollbackTo(name).Error
}

// ReleaseSavepoint Release save point (optional)
func ReleaseSavepoint(ctx context.Context, name string) error {
	tx := GetExecDB(ctx, nil)
	if tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return tx.Exec("RELEASE SAVEPOINT " + name).Error
}
