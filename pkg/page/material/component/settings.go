package component

import (
	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/blocks"
)

type Settings struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	BlockOrder      []string                   `json:"block_order"`
	Blocks          map[string]blocks.Settings `json:"blocks"`
	ElementSettings map[string]jsonx.JSONValue `json:"element_settings"`
}

func (s *Settings) GetElementSettingByID(id string) *jsonx.JSONValue {
	if s.ElementSettings == nil {
		return nil
	}
	val, ok := s.ElementSettings[id]
	if !ok {
		return nil
	}
	return &val
}
