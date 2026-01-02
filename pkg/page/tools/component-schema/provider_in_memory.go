package componentschema

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/page/material/component"
)

type inMemorySchemaProvider struct {
	schemas map[string]component.Schema
}

func NewInMemorySchemaProvider(schemas map[string]component.Schema) ComponentSchemaProvider {
	return &inMemorySchemaProvider{
		schemas: schemas,
	}
}

func (p *inMemorySchemaProvider) Get(name string) (component.Schema, error) {
	schema, ok := p.schemas[name]
	if !ok {
		return component.Schema{}, fmt.Errorf("component schema %s not found", name)
	}
	return schema, nil
}
