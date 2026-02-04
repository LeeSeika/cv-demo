package errs

import "github.com/go-sql-driver/mysql"

type mysqlDBError struct {
	mysqlErr *mysql.MySQLError
}

func (e *mysqlDBError) DetailedErrorMessage() string {
	return e.mysqlErr.Message
}

func (e *mysqlDBError) ConstraintName() string {
	panic("not implemented")
}
