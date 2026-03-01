package template

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (t *template) ForceSaveData(ctx context.Context, id string, data json.RawMessage) (int64, error) {
	db := t.db.WithContext(ctx)

	updates := map[string]any{
		"data":    datatypes.JSON(data),
		"version": gorm.Expr("version + 1"),
	}

	result := db.Model(&object.Template{}).
		Where("id = ?", id).
		Updates(updates)
	return result.RowsAffected, result.Error
}
