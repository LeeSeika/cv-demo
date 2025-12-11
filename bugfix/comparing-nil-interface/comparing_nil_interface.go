package comparingnilinterface

import "time"

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

// CatalogObject Catalog database object
type CatalogObject struct {
	ID        string `gorm:"primaryKey;column:id"`
	Name      string
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

// CatalogModel Catalog response model
type CatalogModel struct {
	ID        string
	Name      string
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
	case *CatalogObject:
		return &CatalogModel{
			ID:        obj.ID,
			Name:      obj.Name,
			CreatedAt: obj.CreatedAt,
			UpdatedAt: obj.UpdatedAt,
		}
	default:
		return nil
	}
}
