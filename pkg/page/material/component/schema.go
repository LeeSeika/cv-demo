package component

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/blocks"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element/field"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

type Schema struct {
	Name      string            `json:"name"`
	MaxBlocks *uint8            `json:"max_blocks,omitempty"`
	Blocks    []blocks.Schema   `json:"blocks,omitempty"`
	Elements  []element.Element `json:"elements"`
}

type RawSchema struct {
	Name      field.TranslatableField `json:"name"`
	MaxBlocks *uint8                  `json:"max_blocks,omitempty"`
	Blocks    []blocks.RawSchema      `json:"blocks,omitempty"`
	Elements  []jsonx.JSONValue       `json:"elements"`
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
	blocks := make([]blocks.Schema, 0, len(rs.Blocks))
	for _, rawBlock := range rs.Blocks {
		block, err := rawBlock.Parse(locale, localeProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block schema: %w", err)
		}
		blocks = append(blocks, *block)
	}
	rs.Name.SetLocale(locale, localeProvider)
	return &Schema{
		Name:      rs.Name.String(),
		MaxBlocks: rs.MaxBlocks,
		Blocks:    blocks,
		Elements:  elements,
	}, nil
}
