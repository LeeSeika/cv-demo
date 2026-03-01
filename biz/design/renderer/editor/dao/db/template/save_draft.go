package template

import (
	"encoding/json"

	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) SaveDraft(ctx context.Context, id string, userID string, draft *cache.TemplateDraft) error {
	key := templateDraftKey(id, userID)
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	err = t.kv.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
	if err != nil {
		return err
	}
	return nil
}
