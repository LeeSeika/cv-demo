package blocks

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

type Schema struct {
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Limit    *uint8            `json:"limit,omitempty"`
	Elements []element.Element `json:"elements,omitempty"`
}

type RawSchema struct {
	Type     string                  `json:"type"`
	Name     field.TranslatableField `json:"name"`
	Limit    *uint8                  `json:"limit,omitempty"`
	Elements []jsonx.JSONValue       `json:"elements,omitempty"`
}

func (rs *RawSchema) Parse(
	locale string,
	localeProvider locale.LocaleProvider,
) (*Schema, error) {
	elements := make([]element.Element, 0, len(rs.Elements))
	for _, rawEle := range rs.Elements {
		ele, err := element.UnmarshalElement(rawEle)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal element: %w", err)
		}
		elements = append(elements, ele)
	}
	for _, ele := range elements {
		ele.SetLocale(locale, localeProvider)
		if err := ele.Validate(); err != nil {
			return nil, fmt.Errorf("element %s(%s) validation failed: %w", ele.GetID(), ele.EleType(), err)
		}
	}
	rs.Name.SetLocale(locale, localeProvider)
	return &Schema{
		Name:     rs.Name.String(),
		Limit:    rs.Limit,
		Type:     rs.Type,
		Elements: elements,
	}, nil
}
