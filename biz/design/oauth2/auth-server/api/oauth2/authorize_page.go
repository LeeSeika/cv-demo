package oauth2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	serviceoauth2 "github.com/leeseika/cv-demo/biz/design/oauth2/auth-server/service/oauth2"
	"github.com/leeseika/cv-demo/pkg/model/dto"
)

func AuthorizePage(c *gin.Context) {
	payload := dto.OAuthPayload{
		ClientID:            c.Query("client_id"),
		CodeChallenge:       c.Query("code_challenge"),
		CodeChallengeMethod: c.Query("code_challenge_method"),
		RedirectURI:         c.Query("redirect_uri"),
		ResponseType:        c.Query("response_type"),
		Scope:               c.Query("scope"),
		State:               c.Query("state"),
	}

	html, err := serviceoauth2.Get().AuthorizePage(c, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to render authorize page"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
