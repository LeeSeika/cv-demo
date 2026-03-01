package template

import (
	"context"
	"errors"

	"github.com/dgraph-io/badger/v4"
	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/leeseika/cv-demo/pkg/errs"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (t *template) GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error) {
	// query draft from cache
	draft, err := t.templateDraftCacheDAO.GetDraftByID(ctx, id, userID)
	if err == nil {
		return draft, nil
	}
	// err != nil
	log.Err(err).Msg("failed to get template draft from cache")
	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		return nil, errors.New("template not exists")
	}

	saveDraftToCache := func(draft *cache.TemplateDraft) {
		err := t.templateDraftCacheDAO.SetDraft(ctx, id, userID, draft)
		if err != nil {
			log.Err(err).Msg("failed to save template draft to cache")
		}
	}

	// query draft from kv store
	draft, err = t.templateDraftDAO.GetDraftByID(ctx, id, userID)
	if err == nil {
		saveDraftToCache(draft)
		return draft, nil
	}
	// err != nil
	// log error only when it's not key not found
	if !errors.Is(err, badger.ErrKeyNotFound) {
		log.Err(err).Msg("failed to get template draft from kv store")
	}

	// query template from db
	tpl, err := t.templateDAO.GetByID(ctx, id)
	if err != nil {
		if errs.IsDBError(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not exists")
		}
		log.Err(err).Msg("failed to get template from db")
		return nil, err
	}
	// convert template to draft
	draft = cache.BuildTemplateDraft(tpl, userID)
	saveDraftToCache(draft)

	return draft, nil
}
