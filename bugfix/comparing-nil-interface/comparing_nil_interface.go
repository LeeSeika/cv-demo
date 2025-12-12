package comparingnilinterface

import (
	"reflect"
	"time"
)

// ProductObject Product database object
type ProductObject struct {
	ID        string `gorm:"primaryKey;column:id"`
	Name      string
	Handle    string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ProductModel Product response model
type ProductModel struct {
	ID        string
	Name      string
	Handle    string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ObjectToModel(object any) any {
	if object == nil {
		return nil
	}

	switch obj := object.(type) {
	case *ProductObject:
		return &ProductModel{
			ID:        obj.ID,
			Name:      obj.Name,
			Handle:    obj.Handle,
			Price:     obj.Price,
			CreatedAt: obj.CreatedAt,
			UpdatedAt: obj.UpdatedAt,
		}
	}
	return nil
}

func ObjectToModel_Fixed(object any) any {
	if reflect.TypeOf(object).Kind() == reflect.Ptr && reflect.ValueOf(object).IsNil() {
		return nil
	}

	switch obj := object.(type) {
	case *ProductObject:
		return &ProductModel{
			ID:        obj.ID,
			Name:      obj.Name,
			Handle:    obj.Handle,
			Price:     obj.Price,
			CreatedAt: obj.CreatedAt,
			UpdatedAt: obj.UpdatedAt,
		}
	}
	return nil
}
