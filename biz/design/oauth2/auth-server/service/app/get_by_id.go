package app

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *app) GetByID(ctx context.Context, appID string) (*object.App, error) {
	var app object.App
	err := a.db.First(&app, "id = ?", appID).Error
	if err != nil {
		return nil, err
	}

	return &app, nil
}
