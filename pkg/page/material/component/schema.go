package component

import (
	"fmt"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/page/material/component/blocks"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

type Schema struct {
	Name      string            `json:"name"`
	MaxBlocks *uint8            `json:"max_blocks,omitempty"`
	Blocks    []blocks.Schema   `json:"blocks,omitempty"`
	Elements  []element.Element `json:"elements"`
}

func Parse(
	raw jsonmodel.ComponentSchema,
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
	blockSchemas := make([]blocks.Schema, 0, len(raw.Blocks))
	for _, rawBlock := range raw.Blocks {
		block, err := blocks.Parse(rawBlock, locale, localeProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block schema: %w", err)
		}
		blockSchemas = append(blockSchemas, *block)
	}
	raw.Name.SetLocale(locale, localeProvider)
	return &Schema{
		Name:      raw.Name.String(),
		MaxBlocks: raw.MaxBlocks,
		Blocks:    blockSchemas,
		Elements:  elements,
	}, nil
}
