package template

import (
	"encoding/json"
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component"
	"github.com/leeseika/cv-demo/pkg/page/material/component/blocks"
	componentschema "github.com/leeseika/cv-demo/pkg/page/tools/component-schema"
)

type JSONTemplate struct {
	Name       string                        `json:"name"`
	Components map[string]component.Settings `json:"components"`
	Order      []string                      `json:"order"`

	schemaProvider componentschema.ComponentSchemaProvider `json:"-"`
}

func ParseJSON(
	raw json.RawMessage,
	schemaProvider componentschema.ComponentSchemaProvider,
) (*JSONTemplate, error) {
	if schemaProvider == nil {
		return nil, fmt.Errorf("schema provider is nil")
	}
	var t JSONTemplate
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json template: %w", err)
	}
	t.schemaProvider = schemaProvider
	return &t, nil
}

func (t *JSONTemplate) Validate(
	handlers ...ElementValueHandler,
) error {
	if t.schemaProvider == nil {
		return fmt.Errorf("schema provider is nil")
	}

	validatedComponents := make(map[string]component.Settings)
	validatedComponentOrder := make([]string, 0, len(t.Order))
	for _, compID := range t.Order {
		compSettings, ok := t.Components[compID]
		if !ok {
			continue
		}
		compSchema, err := t.schemaProvider.Get(compSettings.Name)
		if err != nil {
			return fmt.Errorf("failed to get schema for component %s: %w", compSettings.Name, err)
		}
		// handle element settings of component
		for _, ele := range compSchema.Elements {
			eleVal := compSettings.GetElementSettingByID(ele.GetID())
			if eleVal == nil {
				continue
			}
			var err error
			for _, handler := range handlers {
				var handledEleVal jsonx.JSONValue
				handledEleVal, err = handler.Handle(ele, *eleVal, err)
				*eleVal = handledEleVal
			}
			if err != nil {
				return fmt.Errorf("component %s element %s value handling failed: %w", compID, ele.GetID(), err)
			}
		}

		// blocks
		validatedBlocks := make(map[string]blocks.Settings, len(compSettings.Blocks))
		validatedBlockOrder := make([]string, 0, len(compSettings.BlockOrder))

		blockLimit := compSchema.MaxBlocks
		currBlockCount := uint8(0)

		blockSchemaMap := make(map[string]*blocks.Schema, len(compSchema.Blocks))
		blockTypeCounter := make(map[string]uint8, len(compSchema.Blocks))
		// build mapping between block type and block schema
		for _, blockSchema := range compSchema.Blocks {
			blockSchemaMap[blockSchema.Type] = &blockSchema
		}

		for _, blockID := range compSettings.BlockOrder {
			if blockLimit != nil && currBlockCount >= *blockLimit {
				return fmt.Errorf("component %s exceeds max block limit of %d", compID, *blockLimit)
			}
			blockSettings, ok := compSettings.Blocks[blockID]
			if !ok {
				continue
			}
			blockType := blockSettings.Type
			blockSchema, ok := blockSchemaMap[blockType]
			if !ok {
				return fmt.Errorf("schema for block type %s not found in component %s", blockType, compID)
			}

			// enforce block type limit
			if blockSchema.Limit != nil {
				currCount, ok := blockTypeCounter[blockType]
				if ok && currCount >= *blockSchema.Limit {
					return fmt.Errorf("component %s exceeds block type %s limit of %d", compID, blockType, *blockSchema.Limit)
				}
			}

			// handle element settings of blocks
			for _, ele := range blockSchema.Elements {
				eleVal := blockSettings.GetSettingByID(ele.GetID())
				if eleVal == nil {
					continue
				}
				var err error
				for _, handler := range handlers {
					var handledEleVal jsonx.JSONValue
					handledEleVal, err = handler.Handle(ele, *eleVal, err)
					*eleVal = handledEleVal
				}
				if err != nil {
					return fmt.Errorf("component %s block %s element %s value handling failed: %w", compID, blockID, ele.GetID(), err)
				}
			}

			validatedBlocks[blockID] = blockSettings
			validatedBlockOrder = append(validatedBlockOrder, blockID)

			currBlockCount++
			blockTypeCounter[blockType] = blockTypeCounter[blockType] + 1

			compSettings.Blocks = validatedBlocks
			compSettings.BlockOrder = validatedBlockOrder
		}

		validatedComponents[compID] = compSettings
		validatedComponentOrder = append(validatedComponentOrder, compID)
	}

	t.Components = validatedComponents
	t.Order = validatedComponentOrder

	return nil
}

func (t *JSONTemplate) ToProps() (map[string]any, error) {
	if t.schemaProvider == nil {
		return nil, fmt.Errorf("schema provider is nil")
	}

	props := make(map[string]any)
	for _, compID := range t.Order {
		compSettings, ok := t.Components[compID]
		if !ok {
			continue
		}
		schema, err := t.schemaProvider.Get(compSettings.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema for component %s: %w", compSettings.Name, err)
		}

		compElementProps := make(map[string]any)
		for _, ele := range schema.Elements {
			eleVal := compSettings.GetElementSettingByID(ele.GetID())
			if eleVal == nil {
				continue
			}
			liquidVal, err := ele.ToLiquid(*eleVal)
			if err != nil {
				return nil, fmt.Errorf("component %s element %s to liquid failed: %w", compID, ele.GetID(), err)
			}
			compElementProps[ele.GetID()] = liquidVal
		}

		// handle blocks
		blockPropsSlice := make([]map[string]any, 0, len(compSettings.Blocks))
		blockSchemaMap := make(map[string]*blocks.Schema, len(schema.Blocks))
		// build mapping between block type and block schema
		for _, blockSchema := range schema.Blocks {
			blockSchemaMap[blockSchema.Type] = &blockSchema
		}
		for _, blockID := range compSettings.BlockOrder {
			blockSettings, ok := compSettings.Blocks[blockID]
			if !ok {
				continue
			}
			blockSchema, ok := blockSchemaMap[blockSettings.Type]
			if !ok {
				return nil, fmt.Errorf("schema for block type %s not found in component %s", blockSettings.Type, compID)
			}

			blockProps := make(map[string]any)
			blockElements := make(map[string]any)
			for _, ele := range blockSchema.Elements {
				eleVal := blockSettings.GetSettingByID(ele.GetID())
				if eleVal == nil {
					continue
				}
				liquidVal, err := ele.ToLiquid(*eleVal)
				if err != nil {
					return nil, fmt.Errorf("component %s block %s element %s to liquid failed: %w", compID, blockID, ele.GetID(), err)
				}
				blockElements[ele.GetID()] = liquidVal
			}
			blockProps["id"] = blockSettings.Type
			blockProps["settings"] = blockElements

			blockPropsSlice = append(blockPropsSlice, blockProps)
		}

		compProps := make(map[string]any)
		compProps["id"] = compID
		compProps["settings"] = compElementProps
		compProps["blocks"] = blockPropsSlice

		props[compID] = compProps
	}

	return props, nil
}
