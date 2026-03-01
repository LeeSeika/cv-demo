package template

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) BatchSaveDraft(ctx context.Context, drafts []*cache.TemplateDraft) error {
	if len(drafts) == 0 {
		return nil
	}

	return t.kv.Update(func(txn *badger.Txn) error {
		for _, draft := range drafts {
			key := templateDraftKey(draft.ID, draft.UserID)
			data, err := json.Marshal(draft)
			if err != nil {
				return err
			}
			if err := txn.Set([]byte(key), data); err != nil {
				return err
			}
		}
		return nil
	})
}
