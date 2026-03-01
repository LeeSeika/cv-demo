package template

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

var _template *template
var _initTemplateOnce sync.Once

type (
	Template interface {
		GetByID(ctx context.Context, id string) (*object.Template, error)
		SaveDataCAS(ctx context.Context, id string, data json.RawMessage, currVersion int) (int64, error)
		ForceSaveData(ctx context.Context, id string, data json.RawMessage) (int64, error)
	}

	TemplateDraft interface {
		GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error)
		SaveDraft(ctx context.Context, id string, userID string, draft *cache.TemplateDraft) error
		BatchSaveDraft(ctx context.Context, drafts []*cache.TemplateDraft) error
	}

	template struct {
		db *gorm.DB
		kv *badger.DB
	}
)

func GetTemplate() Template {
	return _template
}

func GetTemplateDraft() TemplateDraft {
	return _template
}

func Init() {
	_initTemplateOnce.Do(func() {
		_template = &template{
			db: driver.GetDB(),
			kv: driver.GetBadgerDB(),
		}
	})
}
