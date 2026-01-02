package locale

import (
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/jsonx"
)

type JSONProvider struct {
	jsonValue jsonx.JSONValue
}

func (jp *JSONProvider) Get(contextKey string) jsonx.JSONValue {
	return jp.jsonValue.Get(contextKey)
}

func NewJSONProvider(data json.RawMessage) LocaleProvider {
	return &JSONProvider{
		jsonValue: jsonx.JSONValue{
			RawMessage: json.RawMessage(data),
		},
	}
}
