package account

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/design/2fa/service/account"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterWithInput(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if len(email) == 0 || len(password) == 0 {
		return "", errors.New("email and password are required")
	}

	return accountsvc.Get().Register(ctx, email, password)
}

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	qrCode, err := RegisterWithInput(c, req.Email, req.Password)
	if err != nil {
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"qr_code_base64": qrCode,
	})
}
