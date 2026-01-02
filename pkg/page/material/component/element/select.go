package element

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
	"github.com/osteele/liquid/values"
)

type Select struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Default string                  `json:"default"`
	Label   field.TranslatableField `json:"label"`
	Options []SelectOption          `json:"options"`
}

type SelectOption struct {
	Value string                  `json:"value"`
	Label field.TranslatableField `json:"label"`
}

func (s *Select) GetID() string {
	return s.ID
}

func (s *Select) EleType() ElementType {
	return ElementTypeSelect
}

func (s *Select) GetDefault() jsonx.JSONValue {
	defaultJV, err := jsonx.NewString(s.Default)
	if err == nil {
		return *defaultJV
	}
	// fallback to first option
	if len(s.Options) > 0 {
		fallbackJV, err := jsonx.NewString(s.Options[0].Value)
		if err == nil {
			return *fallbackJV
		}
	}
	return *jsonx.NewEmpty()
}

func (s *Select) Validate() error {
	if len(s.Options) == 0 {
		return fmt.Errorf("select element must have at least one option")
	}
	defaultFound := false
	for _, option := range s.Options {
		if option.Value == s.Default {
			defaultFound = true
			break
		}
	}
	if !defaultFound {
		return fmt.Errorf("default value %s is not a valid option", s.Default)
	}
	return nil
}

func (s *Select) SetLocale(locale string, provider locale.LocaleProvider) {
	s.Label.SetLocale(locale, provider)
	for i := range s.Options {
		s.Options[i].Label.SetLocale(locale, provider)
	}
}

func (s *Select) CheckValue(val jsonx.JSONValue) (jsonx.JSONValue, error) {
	if !val.IsString() {
		return val, fmt.Errorf("value is not a string")
	}

	innerVal := val.String()
	for _, option := range s.Options {
		if option.Value == innerVal {
			return val, nil
		}
	}

	return val, fmt.Errorf("value %s is not a valid option", innerVal)
}

func (s *Select) ToLiquid(val jsonx.JSONValue) (values.Value, error) {
	var err error
	val, err = s.CheckValue(val)
	if err != nil {
		return nil, err
	}

	innerVal := val.String()
	v := values.ValueOf(innerVal)
	return v, nil
}
