package template

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

const draftTTL = 7 * 24 * time.Hour

func (t *template) BatchSaveDraft(ctx context.Context, drafts []*cache.TemplateDraft) error {
	_ = ctx

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

		entry := badger.NewEntry([]byte(key), data).WithTTL(draftTTL)
		if err := txn.SetEntry(entry); err != nil {
			if err == badger.ErrTxnTooBig {
				if commitErr := txn.Commit(); commitErr != nil {
					txn.Discard()
					return commitErr
				}

				txn = t.kv.NewTransaction(true)
				if retryErr := txn.SetEntry(entry); retryErr != nil {
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
