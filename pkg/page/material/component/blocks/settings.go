package blocks

import "github.com/leeseika/cv-demo/pkg/jsonx"

type Settings struct {
	Type            string                     `json:"type"`
	ID              string                     `json:"id"`
	ElementSettings map[string]jsonx.JSONValue `json:"element_settings,omitempty"`
}

func (s *Settings) GetSettingByID(id string) *jsonx.JSONValue {
	if s.ElementSettings == nil {
		return nil
	}
	val, ok := s.ElementSettings[id]
	if !ok {
		return nil
	}
	return &val
}
