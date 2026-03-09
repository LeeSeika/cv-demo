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

	txn := t.kv.NewTransaction(true)

	for _, draft := range drafts {
		key := templateDraftKey(draft.ID, draft.UserID)
		data, err := json.Marshal(draft)
		if err != nil {
			continue
		}

		if err := txn.Set([]byte(key), data); err != nil {
			if err == badger.ErrTxnTooBig {
				if commitErr := txn.Commit(); commitErr != nil {
					txn.Discard()
					return commitErr
				}

				txn = t.kv.NewTransaction(true)
				if retryErr := txn.Set([]byte(key), data); retryErr != nil {
					txn.Discard()
					return retryErr
				}
				continue
			}

			txn.Discard()
			return err
		}
	}

	if err := txn.Commit(); err != nil {
		txn.Discard()
		return err
	}

	return nil
}
