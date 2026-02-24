package json

type OAuthPayload struct {
	AppID               string `json:"app_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
}

type AuthInfo struct {
	AccountID string `json:"account_id"`
	ShopID    string `json:"shop_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	AppID     string `json:"app_id"`
}
