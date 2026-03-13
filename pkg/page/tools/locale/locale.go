package locale

import "github.com/leeseika/cv-demo/pkg/utils/jsonx"

type LocaleProvider interface {
	Get(contextKey string) jsonx.JSONValue
}
