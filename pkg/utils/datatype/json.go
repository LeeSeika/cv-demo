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

type JSON[T any] struct {
	data T
}

func NewJSON[T any](data T) JSON[T] {
	return JSON[T]{
		data: data,
	}
}

// Data return data with generic Type T
func (j JSON[T]) Data() T {
	return j.data
}

// Value return json value, implement driver.Valuer interface
func (j JSON[T]) Value() (driver.Value, error) {
	ba, err := j.MarshalJSON()
	return string(ba), err
}

// Scan scan value into JSON[T], implements sql.Scanner interface
func (j *JSON[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	case *string:
		if v != nil {
			bytes = []byte(*v)
		} else {
			bytes = []byte("null")
		}
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, &j.data)
}

// MarshalJSON to output non base64 encoded []byte
func (j JSON[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.data)
}

// UnmarshalJSON to deserialize []byte
func (j *JSON[T]) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &j.data); err == nil {
		return nil
	}

	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	return json.Unmarshal([]byte(raw), &j.data)
}

// GormDataType gorm common data type
func (JSON[T]) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSON[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js JSON[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := js.MarshalJSON()

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}
