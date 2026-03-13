package errs

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type postgresDBError struct {
	pgErr *pgconn.PgError
}

func (e *postgresDBError) DetailedErrorMessage() string {
	return e.pgErr.Detail
}

func (e *postgresDBError) ConstraintName() string {
	return e.pgErr.ConstraintName
}

// pg db errors
const (
	pgErrUniqueViolation     = pq.ErrorCode("23505")
	pgErrForeignKeyViolation = pq.ErrorCode("23503")
)

func isPostgresDBError(err error, gormErr error) bool {
	var errCode pq.ErrorCode

	errCode, ok := pgErrorCodeFromGORMError(gormErr)
	if !ok {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == string(errCode)
	}
	return false
}

func pgErrorCodeFromGORMError(gormErr error) (pq.ErrorCode, bool) {
	switch gormErr {
	case gorm.ErrDuplicatedKey:
		return pgErrUniqueViolation, true
	case gorm.ErrForeignKeyViolated:
		return pgErrForeignKeyViolation, true
	}

	return "", false
}
