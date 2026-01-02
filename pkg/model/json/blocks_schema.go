package json

import (
	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
)

type BlocksSchema struct {
	Type     string                  `json:"type"`
	Name     field.TranslatableField `json:"name"`
	Limit    *uint8                  `json:"limit,omitempty"`
	Elements []jsonx.JSONValue       `json:"elements,omitempty"`
}
