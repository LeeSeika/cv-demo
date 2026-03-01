package template

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (t *template) GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error) {
	key := templateDraftKey(id, userID)
	var draft cache.TemplateDraft
	err := t.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &draft)
		})
	})
	if err != nil {
		return nil, err
	}
	return &draft, nil
}
