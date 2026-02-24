package app

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
)

type (
	App interface {
		List(ctx context.Context) ([]*object.App, error)
		CreateAuthSession(ctx context.Context, clientID string) (string, error)
	}

	app struct {
		db *gorm.DB
	}
)
