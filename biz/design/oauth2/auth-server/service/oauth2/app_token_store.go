package oauth2

import (
	"context"
	"encoding/json"
	"fmt"

	oa "github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/app"
	"github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/shop"
	"github.com/leeseika/cv-demo/biz/design/oauth2/pkg/auth"
	"github.com/leeseika/cv-demo/pkg/driver"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func newAppAccessTokenStore(oauth2 OAuth2) (oa.TokenStore, error) {
	return &appAccessTokenStore{
		oauth2: oauth2,
		db:     driver.GetDB(),
	}, nil
}

type (
	appAccessTokenStore struct {
		oauth2 OAuth2
		db     *gorm.DB
	}
)

// Create create and store the new token information
func (t *appAccessTokenStore) Create(ctx context.Context, info oa.TokenInfo) error {
	appID := info.GetClientID()
	shopID := info.GetUserID()

	// we will check the shop status first
	shopStatusIsAvailable, err := shop.Get().VerifyShopAvailability(ctx, shopID)
	if err != nil {
		return fmt.Errorf("failed to verify shop availability: %w", err)
	}
	if !shopStatusIsAvailable {
		return fmt.Errorf("shop is not available")
	}

	oauth2Token := dto.OAuth2Token{}

	oauth2Token.SetClientID(info.GetClientID())
	oauth2Token.SetUserID(info.GetUserID())
	oauth2Token.SetRedirectURI(info.GetRedirectURI())
	oauth2Token.SetScope(info.GetScope())
	oauth2Token.SetCode(info.GetCode())
	oauth2Token.SetCodeCreateAt(info.GetCodeCreateAt())
	oauth2Token.SetCodeExpiresIn(info.GetCodeExpiresIn())
	oauth2Token.SetCodeChallenge(info.GetCodeChallenge())
	oauth2Token.SetCodeChallengeMethod(info.GetCodeChallengeMethod())
	oauth2Token.SetAccess(info.GetAccess())
	oauth2Token.SetAccessCreateAt(info.GetAccessCreateAt())
	oauth2Token.SetAccessExpiresIn(info.GetAccessExpiresIn())
	oauth2Token.SetRefresh(info.GetRefresh())
	oauth2Token.SetRefreshCreateAt(info.GetRefreshCreateAt())
	oauth2Token.SetRefreshExpiresIn(info.GetRefreshExpiresIn())

	oauth2TokenInJSON, err := json.Marshal(oauth2Token)

	if err != nil {
		return err
	}

	var shopApp object.ShopApp
	err = t.db.First(&shopApp, "app_id = ? AND shop_id = ?", appID, shopID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err != nil {
		app, err := app.Get().GetByID(ctx, appID)

		if err != nil {
			log.Error().Msgf("AppAccessTokenStore cannot get application info: %v", err)
			return fmt.Errorf("app not found")
		}

		if app == nil {
			return fmt.Errorf("app not found")
		}

		newShopApp := object.ShopApp{
			AppID:        appID,
			ShopID:       shopID,
			Status:       "in_progress",
			AuthCode:     oauth2Token.Code,
			AccessToken:  oauth2Token.Access,
			RefreshToken: oauth2Token.Refresh,
		}

		err = t.db.Create(&newShopApp).Error
		if err != nil {
			log.Error().Msgf("AppAccessTokenStore cannot insert shop_apps: %v", err)
			return err
		}

		// update shop_apps details with `newShopApp`
		shopApp = newShopApp
	}

	if code := info.GetCode(); code != "" {
		shopApp.AuthCode = code
	} else {
		shopApp.AuthCode = "-"
	}

	if access := info.GetAccess(); access != "" {
		shopApp.AccessToken = access
	}

	if refresh := info.GetRefresh(); refresh != "" {
		shopApp.RefreshToken = refresh
	}

	// update token info
	shopApp.TokenInfo = oauth2TokenInJSON
	err = t.db.Model(&object.ShopApp{}).Where("app_id = ? AND shop_id = ?", appID, shopID).Updates(shopApp).Error
	if err != nil {
		return err
	}

	return nil
}

// RemoveByCode delete the authorization code
func (t *appAccessTokenStore) RemoveByCode(ctx context.Context, code string) error {
	if len(code) == 0 {
		log.Error().Msgf("AppAccessTokenStore access_code is empty")

		return nil
	}

	var shopApp object.ShopApp
	err := t.db.First(&shopApp, "auth_code = ?", code).Error
	if err != nil {
		log.Error().Msgf("AppAccessTokenStore cannot get shop_app record: %+v", err)
		return err
	}

	shopID := shopApp.ShopID
	appID := shopApp.AppID

	shopApp.AuthCode = "-"

	err = t.db.Model(&object.ShopApp{}).Where("app_id = ? AND shop_id = ?", appID, shopID).Updates(shopApp).Error
	if err != nil {
		return err
	}

	return nil
}

// RemoveByAccess use the access token to delete the token information
func (t *appAccessTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	// OAuth2 service does not implement RemoveByAccess()

	return nil
}

// RemoveByRefresh use the refresh token to delete the token information
func (t *appAccessTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	// OAuth2 service does not implement RemoveByRefresh()

	return nil
}

// GetByCode use the authorization code for token information data
func (t *appAccessTokenStore) GetByCode(ctx context.Context, code string) (oa.TokenInfo, error) {
	if len(code) == 0 {
		return nil, fmt.Errorf("AppAccessTokenStore no shop_apps is found")
	}

	if code == "-" {
		return nil, fmt.Errorf("AppAccessTokenStore no shop_apps is found")
	}

	var shopApp object.ShopApp
	err := t.db.First(&shopApp, "auth_code = ?", code).Error
	if err != nil {
		log.Error().Msgf("AppAccessTokenStore cannot get shop_app record: %+v", err)
		return nil, err
	}

	isAuthInProgress := shopApp.Status == "in_progress"
	isInstalled := shopApp.Status == "installed"

	if !isAuthInProgress && !isInstalled {
		return nil, errors.ErrInvalidAccessToken
	}

	tokenInfo, err := t.toTokenInfo(shopApp.TokenInfo)

	if err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

// GetByAccess use the access token for token information data
func (t *appAccessTokenStore) GetByAccess(ctx context.Context, access string) (oa.TokenInfo, error) {
	claims, authInfo, err := auth.ParseToken(access)

	if err != nil {
		return nil, errors.ErrInvalidAccessToken
	}

	if "app-auth" != claims.Subject {
		return nil, errors.ErrInvalidAccessToken
	}

	shopID := authInfo.ShopID
	appID := authInfo.AppID

	if len(shopID) == 0 || len(appID) == 0 {
		return nil, errors.ErrInvalidAccessToken
	}

	var shopApp object.ShopApp
	err = t.db.First(&shopApp, "app_id = ? AND shop_id = ?", appID, shopID).Error
	if err != nil {
		return nil, err
	}

	isAuthInProgress := shopApp.Status == "in_progress"
	isInstalled := shopApp.Status == "installed"

	if !isAuthInProgress && !isInstalled {
		return nil, errors.ErrInvalidAccessToken
	}

	tokenInfo, err := t.toTokenInfo(shopApp.TokenInfo)

	if err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

// GetByRefresh use the refresh token for token information data
func (t *appAccessTokenStore) GetByRefresh(ctx context.Context, refresh string) (oa.TokenInfo, error) {
	return nil, nil
}

func (t *appAccessTokenStore) toTokenInfo(data []byte) (oa.TokenInfo, error) {
	var token dto.OAuth2Token

	err := json.Unmarshal(data, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
