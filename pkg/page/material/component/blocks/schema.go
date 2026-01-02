package blocks

import (
	"fmt"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

type Schema struct {
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Limit    *uint8            `json:"limit,omitempty"`
	Elements []element.Element `json:"elements,omitempty"`
}

func Parse(
	raw jsonmodel.BlocksSchema,
	locale string,
	localeProvider locale.LocaleProvider,
) (*Schema, error) {
	elements := make([]element.Element, 0, len(raw.Elements))
	for _, rawEle := range raw.Elements {
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
	raw.Name.SetLocale(locale, localeProvider)
	return &Schema{
		Name:     raw.Name.String(),
		Limit:    raw.Limit,
		Type:     raw.Type,
		Elements: elements,
	}, nil
}
