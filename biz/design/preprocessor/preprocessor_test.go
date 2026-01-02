package preprocessor

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/leeseika/cv-demo/pkg/page/material/component"
	componentschema "github.com/leeseika/cv-demo/pkg/page/tools/component-schema"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
	"github.com/microcosm-cc/bluemonday"
)

var (
	productPageTemplateRaw      []byte
	localeEnUSRaw               []byte
	localeZhCNRaw               []byte
	productTitleSchemaRaw       []byte
	productDescriptionSchemaRaw []byte
)

func init() {
	var err error
	productPageTemplateRaw, err = os.ReadFile("./test-data/template/product_page.json")
	if err != nil {
		panic(err)
	}
	localeEnUSRaw, err = os.ReadFile("./test-data/locale/en-US.json")
	if err != nil {
		panic(err)
	}
	localeZhCNRaw, err = os.ReadFile("./test-data/locale/zh-CN.json")
	if err != nil {
		panic(err)
	}
	productTitleSchemaRaw, err = os.ReadFile("./test-data/component-schema/product_title.json")
	if err != nil {
		panic(err)
	}
	productDescriptionSchemaRaw, err = os.ReadFile("./test-data/component-schema/product_description.json")
	if err != nil {
		panic(err)
	}
}

func TestPreprocessProductPage(t *testing.T) {
	enProvider := locale.NewJSONProvider(localeEnUSRaw)
	zhProvider := locale.NewJSONProvider(localeZhCNRaw)

	tests := []struct {
		name           string
		locale         string
		localeProvider locale.LocaleProvider
	}{
		{
			name:           "en-US",
			locale:         "en-US",
			localeProvider: enProvider,
		},
		{
			name:           "zh-CN",
			locale:         "zh-CN",
			localeProvider: zhProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// preprocess component schemas
			// parse raw component schemas
			productTitleComponentSchema, err := PreprocessComponent(
				productTitleSchemaRaw,
				tt.locale,
				tt.localeProvider,
			)
			if err != nil {
				t.Fatalf("failed to handle product title component schema: %v", err)
			}
			productDescriptionComponentSchema, err := PreprocessComponent(
				productDescriptionSchemaRaw,
				tt.locale,
				tt.localeProvider,
			)
			if err != nil {
				t.Fatalf("failed to handle product description component schema: %v", err)
			}
			// prepare component schema provider
			schemaMap := map[string]component.Schema{
				"product_title":       *productTitleComponentSchema,
				"product_description": *productDescriptionComponentSchema,
			}
			componentSchemaProvider := componentschema.NewInMemorySchemaProvider(schemaMap)

			parsedProductTitleSchema, _ := json.Marshal(productTitleComponentSchema)
			parsedProductDescriptionSchema, _ := json.Marshal(productDescriptionComponentSchema)

			t.Logf("preprocessed component schema %s with locale %v: %v",
				productTitleComponentSchema.Name,
				tt.locale,
				string(parsedProductTitleSchema),
			)
			t.Logf("preprocessed component schema %s with locale %v: %v",
				productDescriptionComponentSchema.Name,
				tt.locale,
				string(parsedProductDescriptionSchema),
			)

			// preprocess product page template
			// parse and validate JSON template
			_, err = PreprocessJSONTemplate(
				productPageTemplateRaw,
				componentSchemaProvider,
				NewElementValueSanitizer(bluemonday.UGCPolicy()),
			)
			if err != nil {
				t.Fatalf("failed to handle product page template: %v", err)
			}

			// t.Logf("preprocessed product page template with locale %v: %+v", tt.locale, jsonTpl)
		})
	}
}
