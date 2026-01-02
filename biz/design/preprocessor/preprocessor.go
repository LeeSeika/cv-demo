package preprocessor

import (
	"encoding/json"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/page/material/component"
	"github.com/leeseika/cv-demo/pkg/page/material/template"
	componentschema "github.com/leeseika/cv-demo/pkg/page/tools/component-schema"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

func PreprocessComponent(
	raw json.RawMessage,
	locale string,
	localeProvider locale.LocaleProvider,
) (*component.Schema, error) {
	var rawComponentSchema jsonmodel.ComponentSchema
	if err := json.Unmarshal(raw, &rawComponentSchema); err != nil {
		return nil, err
	}
	componentSchema, err := component.Parse(rawComponentSchema, locale, localeProvider)
	if err != nil {
		return nil, err
	}
	return componentSchema, nil
}

func PreprocessJSONTemplate(
	raw json.RawMessage,
	schemaProvider componentschema.ComponentSchemaProvider,
	additionalHandlers ...template.ElementValueHandler,
) (*template.JSONTemplate, error) {
	tpl, err := template.ParseJSON(raw, schemaProvider)
	if err != nil {
		return nil, err
	}

	handlers := []template.ElementValueHandler{
		// built-in handlers
		template.NewElementValueChecker(),
		template.NewElementValueDefaultSetter(),
	}
	// append additional handlers
	handlers = append(handlers, additionalHandlers...)

	err = tpl.Validate(handlers...)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
