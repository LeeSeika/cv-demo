package template

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/datatypes"
)

func (t *template) SaveDataCAS(ctx context.Context, id string, data json.RawMessage, currVersion int) (int64, error) {
	db := t.db.WithContext(ctx)

	updatedTemplate := object.Template{
		Data:    datatypes.JSON(data),
		Version: currVersion + 1,
	}

	result := db.Model(&object.Template{}).
		Where("id = ? AND version = ?", id, currVersion).
		Updates(updatedTemplate)
	return result.RowsAffected, result.Error
}
