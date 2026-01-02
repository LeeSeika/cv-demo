package locale

import "github.com/leeseika/cv-demo/pkg/jsonx"

type LocaleProvider interface {
	Get(contextKey string) jsonx.JSONValue
}
