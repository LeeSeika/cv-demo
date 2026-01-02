package datatype

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type JSONSlice[T any] []T

func NewJSONSlice[T any](s []T) JSONSlice[T] {
	return JSONSlice[T](s)
}

// Value return json value, implement driver.Valuer interface
func (j JSONSlice[T]) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan scan value into JSON[T], implements sql.Scanner interface
func (j *JSONSlice[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, &j)
}

// GormDataType gorm common data type
func (JSONSlice[T]) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSONSlice[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (j JSONSlice[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	switch db.Dialector.Name() {
	case "mysql":
		data, _ := json.Marshal(j)
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	data, _ := json.Marshal(j)
	return gorm.Expr("?", string(data))
}

func (j JSONSlice[T]) String() string {
	bytes, _ := json.Marshal(j)
	return string(bytes)
}
