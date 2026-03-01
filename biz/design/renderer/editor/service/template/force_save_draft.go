package template

import (
	"context"
	"errors"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (t *template) ForceSaveDraft(ctx context.Context, id string, draft *dto.ForceSaveTemplateDraftReq) error {
	rowsAffected, err := t.templateDAO.ForceSaveData(ctx, id, draft.Data)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("template not exists")
	}
	return nil
}
