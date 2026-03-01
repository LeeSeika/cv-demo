package json

type AuthInfo struct {
	AccountID string `json:"account_id"`
	ShopID    string `json:"shop_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	AppID     string `json:"app_id"`
}

type OAuth2AppInfo struct {
	ID     string
	Secret string
	Domain string
	UserID string
}

// IsPublic implements [oauth2.ClientInfo].
func (ai *OAuth2AppInfo) IsPublic() bool {
	return true
}

// GetID client id
func (ai *OAuth2AppInfo) GetID() string {
	return ai.ID
}

// GetSecret client secret
func (ai *OAuth2AppInfo) GetSecret() string {
	return ai.Secret
}

// GetDomain client domain
func (ai *OAuth2AppInfo) GetDomain() string {
	return ai.Domain
}

// GetUserID user id
func (ai *OAuth2AppInfo) GetUserID() string {
	return ai.UserID
}
