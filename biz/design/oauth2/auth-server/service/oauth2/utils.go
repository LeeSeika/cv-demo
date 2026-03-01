package oauth2

import (
	"fmt"
	"net/http"

	"github.com/go-session/session/v3"
	"github.com/leeseika/cv-demo/biz/design/oauth2/pkg/constants"
)

var authHandler = func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	store, err := session.Start(r.Context(), w, r)

	if err != nil {
		return
	}

	uid, ok := store.Get(constants.AppAuthSessionNameKey)

	if !ok {
		w.Header().Set("Location", "/auth/login")
		w.WriteHeader(http.StatusFound)

		return
	}

	userID, ok = uid.(string)

	if !ok {
		return "", fmt.Errorf("no user id is found")
	}

	store.Delete(constants.AppAuthSessionNameKey)
	store.Delete(constants.AppAuthSessionHashKey)
	store.Save()

	return
}
