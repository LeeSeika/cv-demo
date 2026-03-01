package dto

import (
	"time"

	oa "github.com/go-oauth2/oauth2/v4"
)

type OAuthPayload struct {
	ClientID            string `json:"client_id" binding:"required"`
	CodeChallenge       string `json:"code_challenge" binding:"required"`
	CodeChallengeMethod string `json:"code_challenge_method" binding:"required"`
	RedirectURI         string `json:"redirect_uri" binding:"required"`
	ResponseType        string `json:"response_type" binding:"required"`
	Scope               string `json:"scope" binding:"required"`
	State               string `json:"state" binding:"required"`
}

type AuthorizeRequest struct {
	OAuthPayload
	CSRFToken string `json:"csrf_token" binding:"required"`
}

// OAuth2Token token model
type OAuth2Token struct {
	ClientID            string        `json:"client_id"`
	UserID              string        `json:"user_id"`
	RedirectURI         string        `json:"redirect_uri"`
	Scope               string        `json:"scope"`
	Code                string        `json:"code"`
	CodeChallenge       string        `json:"code_challenge"`
	CodeChallengeMethod string        `json:"code_challenge_method"`
	CodeCreateAt        time.Time     `json:"code_create_at"`
	CodeExpiresIn       time.Duration `json:"code_expires_in"`
	Access              string        `json:"access"`
	AccessCreateAt      time.Time     `json:"access_create_at"`
	AccessExpiresIn     time.Duration `json:"access_expires_in"`
	Refresh             string        `json:"refresh"`
	RefreshCreateAt     time.Time     `json:"refresh_create_at"`
	RefreshExpiresIn    time.Duration `json:"refresh_expires_in"`
}

// New implements [oauth2.TokenInfo].
func (t *OAuth2Token) New() oa.TokenInfo {
	return &OAuth2Token{}
}

// GetClientID the client id
func (t *OAuth2Token) GetClientID() string {
	return t.ClientID
}

// SetClientID the client id
func (t *OAuth2Token) SetClientID(clientID string) {
	t.ClientID = clientID
}

// GetUserID the user id
func (t *OAuth2Token) GetUserID() string {
	return t.UserID
}

// SetUserID the user id
func (t *OAuth2Token) SetUserID(userID string) {
	t.UserID = userID
}

// GetRedirectURI redirect URI
func (t *OAuth2Token) GetRedirectURI() string {
	return t.RedirectURI
}

// SetRedirectURI redirect URI
func (t *OAuth2Token) SetRedirectURI(redirectURI string) {
	t.RedirectURI = redirectURI
}

// GetScope get scope of authorization
func (t *OAuth2Token) GetScope() string {
	return t.Scope
}

// SetScope get scope of authorization
func (t *OAuth2Token) SetScope(scope string) {
	t.Scope = scope
}

// GetCode authorization code
func (t *OAuth2Token) GetCode() string {
	return t.Code
}

// SetCode authorization code
func (t *OAuth2Token) SetCode(code string) {
	t.Code = code
}

// GetCodeCreateAt create Time
func (t *OAuth2Token) GetCodeCreateAt() time.Time {
	return t.CodeCreateAt
}

// SetCodeCreateAt create Time
func (t *OAuth2Token) SetCodeCreateAt(createAt time.Time) {
	t.CodeCreateAt = createAt
}

// GetCodeExpiresIn the lifetime in seconds of the authorization code
func (t *OAuth2Token) GetCodeExpiresIn() time.Duration {
	return t.CodeExpiresIn
}

// SetCodeExpiresIn the lifetime in seconds of the authorization code
func (t *OAuth2Token) SetCodeExpiresIn(exp time.Duration) {
	t.CodeExpiresIn = exp
}

// GetCodeChallenge challenge code
func (t *OAuth2Token) GetCodeChallenge() string {
	return t.CodeChallenge
}

// SetCodeChallenge challenge code
func (t *OAuth2Token) SetCodeChallenge(code string) {
	t.CodeChallenge = code
}

// GetCodeChallengeMethod challenge method
func (t *OAuth2Token) GetCodeChallengeMethod() oa.CodeChallengeMethod {
	return oa.CodeChallengeMethod(t.CodeChallengeMethod)
}

// SetCodeChallengeMethod challenge method
func (t *OAuth2Token) SetCodeChallengeMethod(method oa.CodeChallengeMethod) {
	t.CodeChallengeMethod = string(method)
}

// GetAccess access Token
func (t *OAuth2Token) GetAccess() string {
	return t.Access
}

// SetAccess access Token
func (t *OAuth2Token) SetAccess(access string) {
	t.Access = access
}

// GetAccessCreateAt create Time
func (t *OAuth2Token) GetAccessCreateAt() time.Time {
	return t.AccessCreateAt
}

// SetAccessCreateAt create Time
func (t *OAuth2Token) SetAccessCreateAt(createAt time.Time) {
	t.AccessCreateAt = createAt
}

// GetAccessExpiresIn the lifetime in seconds of the access token
func (t *OAuth2Token) GetAccessExpiresIn() time.Duration {
	return t.AccessExpiresIn
}

// SetAccessExpiresIn the lifetime in seconds of the access token
func (t *OAuth2Token) SetAccessExpiresIn(exp time.Duration) {
	t.AccessExpiresIn = exp
}

// GetRefresh refresh Token
func (t *OAuth2Token) GetRefresh() string {
	return t.Refresh
}

// SetRefresh refresh Token
func (t *OAuth2Token) SetRefresh(refresh string) {
	t.Refresh = refresh
}

// GetRefreshCreateAt create Time
func (t *OAuth2Token) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAt
}

// SetRefreshCreateAt create Time
func (t *OAuth2Token) SetRefreshCreateAt(createAt time.Time) {
	t.RefreshCreateAt = createAt
}

// GetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *OAuth2Token) GetRefreshExpiresIn() time.Duration {
	return t.RefreshExpiresIn
}

// SetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *OAuth2Token) SetRefreshExpiresIn(exp time.Duration) {
	t.RefreshExpiresIn = exp
}
