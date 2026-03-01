package template

import (
	"context"
	"sync"

	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
)

var _template *template
var _initTemplateOnce sync.Once

type (
	TemplateDraft interface {
		GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error)
		GetDraftByKey(ctx context.Context, key string) (*cache.TemplateDraft, error)
		GetMultiDraftsByKeys(ctx context.Context, keys []string) (map[string]*cache.TemplateDraft, error)
		SetDraft(ctx context.Context, id string, userID string, draft *cache.TemplateDraft) error
	}

	template struct {
		cache driver.KVCacheProvider
	}
)

func InitTemplateDraft() {
	_initTemplateOnce.Do(func() {
		_template = &template{
			cache: driver.GetKVCacheProvider(),
		}
	})
}

func GetTemplateDraft() TemplateDraft {
	return _template
}
