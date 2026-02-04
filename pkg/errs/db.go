package errs

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/leeseika/cv-demo/pkg/driver"
)

type GenericDBError interface {
	DetailedErrorMessage() string
	ConstraintName() string
}

// IsDBError checks if the given error matches the specified gorm error, considering the underlying database dialect.
func IsDBError(err error, gormErr error) bool {
	if errors.Is(err, gormErr) {
		return true
	}

	db := driver.GetDB()
	dialector := db.Dialector.Name()

	switch dialector {
	case "postgres":
		return isPostgresDBError(err, gormErr)

	case "mysql":
		panic("not implemented")

	default:
		return isPostgresDBError(err, gormErr)
	}
}

func MustUnwrapDBError(err error) GenericDBError {
	db := driver.GetDB()
	dialector := db.Dialector.Name()

	switch dialector {
	case "postgres":
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return &postgresDBError{pgErr: pgErr}
		} else {
			panic("unexpected error type")
		}
	case "mysql":
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			return &mysqlDBError{mysqlErr: mySQLErr}
		} else {
			panic("unexpected error type")
		}

	default:
		panic(fmt.Sprintf("unsupported dialector %s", dialector))
	}
}
