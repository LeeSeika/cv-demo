/*
Copyright: 2023, Deep Codify Limited

auth-api server
*/

package oauth2

import (
	"context"
	"fmt"

	oa "github.com/go-oauth2/oauth2/v4"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/app"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/rs/zerolog/log"
)

// newAppInfoStore returns App Info Storage
func newAppInfoStore(oauth2 OAuth2) oa.ClientStore {
	return &appInfoStore{
		oauth2: oauth2,
	}
}

type (
	// appInfoStore Application Information Storage
	appInfoStore struct {
		oauth2 OAuth2
	}
)

// GetByID get client info by ID
func (cs *appInfoStore) GetByID(ctx context.Context, id string) (oa.ClientInfo, error) {
	app, err := app.Get().GetByID(ctx, id)

	if err != nil {
		log.Error().Msgf("cannot get application info: %v", err)
		return nil, fmt.Errorf("not found")
	}

	appInfo := newAppInfo(app.ID, app.Secret, app.RedirectURL, "")

	return &appInfo, nil
}

func newAppInfo(id, secret, domain, userID string) jsonmodel.OAuth2AppInfo {
	return jsonmodel.OAuth2AppInfo{
		ID:     id,
		Secret: secret,
		Domain: domain,
		UserID: userID,
	}
}
