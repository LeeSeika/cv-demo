package dto

import (
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

type ForceSaveTemplateDraftReq struct {
	Data json.RawMessage `json:"data"`
}

type SaveTemplateDraftReq struct {
	Data    json.RawMessage `json:"data"`
	Version int             `json:"version"`
}

type EditTemplateDraftReq struct {
	cache.TemplateDraft
}
