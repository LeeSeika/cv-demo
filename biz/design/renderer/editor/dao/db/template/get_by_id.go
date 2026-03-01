package template

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (t *template) GetByID(ctx context.Context, id string) (*object.Template, error) {
	db := t.db.WithContext(ctx)

	var tmpl object.Template
	if err := db.First(&tmpl, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}
