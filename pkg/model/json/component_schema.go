package json

import (
	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
)

type ComponentSchema struct {
	Name      field.TranslatableField `json:"name"`
	MaxBlocks *uint8                  `json:"max_blocks,omitempty"`
	Blocks    []BlocksSchema          `json:"blocks,omitempty"`
	Elements  []jsonx.JSONValue       `json:"elements"`
}
