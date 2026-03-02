package account

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/2fa/service/account"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginWithInput(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if len(email) == 0 || len(password) == 0 {
		return "", errors.New("email and password are required")
	}

	return accountsvc.Get().Login(ctx, email, password)
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	verificationToken, err := LoginWithInput(c, req.Email, req.Password)
	if err != nil {
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verification_token": verificationToken,
	})
}
