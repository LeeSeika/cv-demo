package oauth2

import (
	"context"
	"errors"

	"github.com/go-oauth2/oauth2/v4/server"
)

func (o *oauth2) ValidateCSRFToken(ctx context.Context, csrfToken string, hashed string, authorizeRequest *server.AuthorizeRequest) error {
	raw := csrfToken + authorizeRequest.ClientID + authorizeRequest.CodeChallenge + string(authorizeRequest.CodeChallengeMethod) + authorizeRequest.RedirectURI + string(authorizeRequest.ResponseType) + authorizeRequest.Scope + authorizeRequest.State
	hashedRaw := generateHash(raw)

	if hashed != hashedRaw {
		return errors.New("invalid CSRF token")
	}

	return nil
}
