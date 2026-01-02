package element

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
	"github.com/osteele/liquid/values"
)

type Text struct {
	ID          string                   `json:"id"`
	Type        string                   `json:"type"`
	Default     field.TranslatableField  `json:"default"`
	Label       field.TranslatableField  `json:"label"`
	Placeholder *field.TranslatableField `json:"placeholder,omitempty"`
	WordLimit   *int64                   `json:"word_limit,omitempty"`
}

func (t *Text) GetID() string {
	return t.ID
}

func (t *Text) EleType() ElementType {
	return ElementTypeText
}

func (t *Text) GetDefault() jsonx.JSONValue {
	return t.Default.JSONValue
}

func (t *Text) Validate() error {
	if t.WordLimit != nil && *t.WordLimit <= 0 {
		return fmt.Errorf("word limit must be positive")
	}
	return nil
}

func (t *Text) SetLocale(locale string, provider locale.LocaleProvider) {
	if t.Placeholder != nil {
		t.Placeholder.SetLocale(locale, provider)
	}
	t.Label.SetLocale(locale, provider)
	t.Default.SetLocale(locale, provider)
}

func (t *Text) CheckValue(val jsonx.JSONValue) (jsonx.JSONValue, error) {
	if !val.IsString() {
		return val, fmt.Errorf("value is not a string")
	}

	if t.WordLimit != nil {
		innerVal := val.String()
		if int64(len(innerVal)) > *t.WordLimit {
			innerVal = innerVal[:*t.WordLimit]
			trimmedStr, err := jsonx.NewString(innerVal)
			if err != nil {
				return val, fmt.Errorf("failed to trim string to word limit: %w", err)
			}
			return *trimmedStr, nil
		}
	}

	return val, nil
}

func (t *Text) ToLiquid(val jsonx.JSONValue) (values.Value, error) {
	var err error
	val, err = t.CheckValue(val)
	if err != nil {
		return nil, err
	}

	innerVal := val.String()
	v := values.ValueOf(innerVal)
	return v, nil
}
