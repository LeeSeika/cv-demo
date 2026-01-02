package element

import (
	"encoding/json"
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
	"github.com/osteele/liquid/values"
)

type ElementType string

const (
	ElementTypeRange  ElementType = "range"
	ElementTypeText   ElementType = "text"
	ElementTypeSelect ElementType = "select"
)

type Element interface {
	GetID() string
	EleType() ElementType
	Validate() error
	SetLocale(locale string, provider locale.LocaleProvider)
	CheckValue(val jsonx.JSONValue) (jsonx.JSONValue, error)
	GetDefault() jsonx.JSONValue
	ToLiquid(val jsonx.JSONValue) (values.Value, error)
}

func UnmarshalElement(
	rawEle jsonx.JSONValue,
) (Element, error) {
	eleTypeJV := rawEle.Get("type")
	if !eleTypeJV.IsString() {
		return nil, fmt.Errorf("unexpected element jsonx type %s", eleTypeJV.Result().Type)
	}

	eleType := eleTypeJV.String()
	switch ElementType(eleType) {
	case ElementTypeText:
		var text Text
		if err := json.Unmarshal(rawEle.RawMessage, &text); err != nil {
			return nil, err
		}
		return &text, nil
	case ElementTypeRange:
		var rg Range
		if err := json.Unmarshal(rawEle.RawMessage, &rg); err != nil {
			return nil, err
		}
		return &rg, nil
	case ElementTypeSelect:
		var sel Select
		if err := json.Unmarshal(rawEle.RawMessage, &sel); err != nil {
			return nil, err
		}
		return &sel, nil
	default:
		return nil, fmt.Errorf("unsupported element type %s", eleType)
	}
}
