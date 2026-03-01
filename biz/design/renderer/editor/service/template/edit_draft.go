package template

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (t *template) EditDraft(ctx context.Context, id string, userID string, req *dto.EditTemplateDraftReq) error {
	// save draft to cache
	err := t.templateDraftCacheDAO.SetDraft(ctx, id, userID, &req.TemplateDraft)
	if err != nil {
		return err
	}
	return nil
}
