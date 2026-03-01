package template

import (
	"context"
	"sync"

	templateCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/editor/dao/cache/template"
	templateDAO "github.com/leeseika/cv-demo/biz/design/renderer/editor/dao/db/template"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/leeseika/cv-demo/pkg/model/dto"
)

var _template Template
var _initTemplateOnce sync.Once

type (
	Template interface {
		GetDraftByID(ctx context.Context, id string, userID string) (*cache.TemplateDraft, error)
		EditDraft(ctx context.Context, id string, userID string, draft *dto.EditTemplateDraftReq) error
		TrySaveDraft(ctx context.Context, id string, draft *dto.SaveTemplateDraftReq) error
		ForceSaveDraft(ctx context.Context, id string, draft *dto.ForceSaveTemplateDraftReq) error
	}

	template struct {
		templateDAO           templateDAO.Template
		templateDraftCacheDAO templateCacheDAO.TemplateDraft
		templateDraftDAO      templateDAO.TemplateDraft
	}
)

func Init(templateDAO templateDAO.Template, templateDraftCacheDAO templateCacheDAO.TemplateDraft, templateDraftDAO templateDAO.TemplateDraft) {
	_initTemplateOnce.Do(func() {
		_template = &template{
			templateDAO:           templateDAO,
			templateDraftCacheDAO: templateDraftCacheDAO,
			templateDraftDAO:      templateDraftDAO,
		}
	})
}

func Get() Template {
	return _template
}
