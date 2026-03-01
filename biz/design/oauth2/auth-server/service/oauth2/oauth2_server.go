package oauth2

import "github.com/go-oauth2/oauth2/v4/server"

func (o *oauth2) OAuth2Server() *server.Server {
	return _srv
}
