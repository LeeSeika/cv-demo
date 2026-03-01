package template

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) GetMultiDraftsByKeys(ctx context.Context, keys []string) (map[string]*cache.TemplateDraft, error) {
	bytes, err := t.cache.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}
	drafts := make(map[string]*cache.TemplateDraft, len(bytes))
	for key, b := range bytes {
		var draft cache.TemplateDraft
		err = json.Unmarshal(b, &draft)
		if err != nil {
			continue
		}
		drafts[key] = &draft
	}

	return drafts, nil
}
