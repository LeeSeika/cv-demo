package template

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) GetDraftByKey(ctx context.Context, key string) (*cache.TemplateDraft, error) {
	b, err := t.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var draft cache.TemplateDraft
	err = json.Unmarshal(b, &draft)
	if err != nil {
		return nil, err
	}

	return &draft, nil
}
