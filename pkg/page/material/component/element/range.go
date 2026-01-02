package element

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
	"github.com/osteele/liquid/values"
)

type Range struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Max     int64                   `json:"max"`
	Min     int64                   `json:"min"`
	Default int64                   `json:"default"`
	Unit    string                  `json:"unit"`
	Label   field.TranslatableField `json:"label"`
}

func (r *Range) GetID() string {
	return r.ID
}

func (r *Range) EleType() ElementType {
	return ElementTypeRange
}

func (r *Range) GetDefault() jsonx.JSONValue {
	return *jsonx.NewNumber(float64(r.Default))
}

func (r *Range) Validate() error {
	if r.Min >= r.Max {
		return fmt.Errorf("range min (%d) must be less than max (%d)", r.Min, r.Max)
	}
	_, err := r.CheckValue(*jsonx.NewNumber(float64(r.Default)))
	return err
}

func (r *Range) SetLocale(locale string, provider locale.LocaleProvider) {
	r.Label.SetLocale(locale, provider)
}

func (r *Range) CheckValue(val jsonx.JSONValue) (jsonx.JSONValue, error) {
	if !val.IsNumber() {
		return val, fmt.Errorf("value %v is not a number", val.Result().Value())
	}

	innerVal := val.Num()
	if innerVal < float64(r.Min) || innerVal > float64(r.Max) {
		return val, fmt.Errorf("value %v out of range [%d, %d]", innerVal, r.Min, r.Max)
	}

	return val, nil
}

func (r *Range) ToLiquid(val jsonx.JSONValue) (values.Value, error) {
	var err error
	val, err = r.CheckValue(val)
	if err != nil {
		return nil, err
	}

	innerVal := val.Num()
	v := values.ValueOf(innerVal)
	return v, nil
}
