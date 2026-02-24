package oauth2

import "context"

type (
	Oauth2 interface {
		AuthorizePage(ctx context.Context) (string, error)
		Authorize(ctx context.Context) error
	}
)
