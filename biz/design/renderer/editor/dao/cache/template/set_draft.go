package template

import (
	"context"
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) SetDraft(ctx context.Context, id string, userID string, draft *cache.TemplateDraft) error {
	key := templateDraftKey(id, userID)

	b, err := json.Marshal(draft)
	if err != nil {
		return err
	}

	return t.cache.Set(ctx, key, b, defaultTTL)
}
