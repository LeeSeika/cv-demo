package field

import (
	"strings"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/tools/locale"
)

type TranslatableField struct {
	jsonx.JSONValue
}

func (t *TranslatableField) SetLocale(locale string, provider locale.LocaleProvider) {
	if t.IsString() {
		if provider == nil {
			return
		}
		labelStr := t.String()
		if !strings.HasPrefix(labelStr, "t:") {
			return
		}
		contextKey := strings.TrimPrefix(labelStr, "t:")
		localizedLabel := provider.Get(contextKey)
		if !localizedLabel.IsString() {
			return
		}
		t.JSONValue = localizedLabel
	} else if t.IsObject() {
		localizedLabel := t.Get(locale)
		if !localizedLabel.IsString() {
			localizedLabel = t.Get("default")
		}
		if !localizedLabel.IsString() {
			return
		}
		t.JSONValue = localizedLabel
	}
}
