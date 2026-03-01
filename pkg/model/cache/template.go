package cache

import (
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

type Template struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Data    json.RawMessage `json:"data"`
	Version int             `json:"version"`
}

type TemplateDraft struct {
	UserID string `json:"user_id"`
	Template
}

func TemplateFromObject(tpl *object.Template) *Template {
	return &Template{
		ID:      tpl.ID,
		Name:    tpl.Name,
		Data:    json.RawMessage(tpl.Data),
		Version: tpl.Version,
	}
}

func BuildTemplateDraft(tpl *object.Template, userID string) *TemplateDraft {
	return &TemplateDraft{
		UserID:   userID,
		Template: *TemplateFromObject(tpl),
	}
}
