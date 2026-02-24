package app

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
)

func (a *app) List(ctx context.Context) ([]*object.App, error) {
	var apps []*object.App
	err := a.db.Find(&apps).Error
	if err != nil {
		return nil, err
	}

	return apps, nil
}
