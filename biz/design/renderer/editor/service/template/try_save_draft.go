package template

import (
	"context"
	"errors"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func (t *template) TrySaveDraft(ctx context.Context, id string, req *dto.SaveTemplateDraftReq) error {
	rowsAffected, err := t.templateDAO.SaveDataCAS(ctx, id, req.Data, req.Version)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("template not exists or version mismatch")
	}
	return nil
}
