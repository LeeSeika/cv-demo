package template

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error) {
	key := templateDraftKey(id, userID)

	return t.GetDraftByKey(ctx, key)
}
