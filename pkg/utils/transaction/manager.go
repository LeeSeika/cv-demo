package transaction

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbgorm"
	"github.com/leeseika/cv-demo/pkg/driver"
	"gorm.io/gorm"
)

// TxHandler Managing transaction lifecycle and context propagation
type TxHandler struct {
	db *gorm.DB
}

// NewHandler Creating a Transaction Processor
func NewHandler(ctx context.Context) *TxHandler {
	db := driver.GetDB()
	return &TxHandler{
		db: db,
	}
}

// Run Execute transaction operation (formerly known as ExecuteTx)
// Parameter description:
// - ctx: context, used for timeout control
// - opts: transaction isolation level options
// - operation: business operation function that requires a transaction
func (h *TxHandler) Run(ctx context.Context, opts *sql.TxOptions, operation func(ctx context.Context) error) error {
	return crdbgorm.ExecuteTx(ctx, h.db, opts, func(tx *gorm.DB) error {
		// Injecting a transaction instance into the context
		ctxWithTx := WithTx(ctx, tx)
		// Execute business actions and pass enhanced context
		return operation(ctxWithTx)
	})
}
