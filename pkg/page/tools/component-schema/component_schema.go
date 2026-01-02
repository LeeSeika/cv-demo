package componentschema

import "github.com/leeseika/cv-demo/pkg/page/material/component"

type ComponentSchemaProvider interface {
	Get(name string) (component.Schema, error)
}
