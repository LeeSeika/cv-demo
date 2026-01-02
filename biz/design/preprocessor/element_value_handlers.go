package preprocessor

import (
	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
	"github.com/leeseika/cv-demo/pkg/page/material/template"
)

type Sanitizer interface {
	Sanitize(value string) string
}

type elementValueSanitizer struct {
	sanitizer Sanitizer
}

func NewElementValueSanitizer(sanitizer Sanitizer) template.ElementValueHandler {
	return &elementValueSanitizer{
		sanitizer: sanitizer,
	}
}

func (evs *elementValueSanitizer) Handle(ele element.Element, val jsonx.JSONValue, prevErr error) (jsonx.JSONValue, error) {
	if prevErr != nil {
		return val, prevErr
	}

	// pass through non-string values
	if !val.IsString() {
		return val, nil
	}

	valStr := val.String()
	sanitizedStr := evs.sanitizer.Sanitize(valStr)
	sanitizedJV, err := jsonx.NewString(sanitizedStr)
	if err != nil {
		return val, err
	}

	return *sanitizedJV, nil
}
