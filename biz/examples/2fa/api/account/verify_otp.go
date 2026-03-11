package account

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	accountsvc "github.com/leeseika/cv-demo/biz/examples/2fa/service/account"
	authsvc "github.com/leeseika/cv-demo/biz/examples/2fa/service/auth"
	"github.com/leeseika/cv-demo/pkg/constants"
)

type verifyOTPRequest struct {
	OTP               string `json:"otp" binding:"required"`
	VerificationToken string `json:"verification_token"`
}

func VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token := req.VerificationToken
	if len(token) == 0 {
		token = bearerToken(c)
	}
	token = strings.TrimSpace(token)
	otp := strings.TrimSpace(req.OTP)
	if len(token) == 0 {
		err := errors.New("verification token required")
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}
	if len(otp) == 0 {
		err := errors.New("otp is required")
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	expired, info, subject, err := authsvc.Get().CheckToken(token)
	if err != nil || expired || subject != string(constants.JWTSubjectTOTPVerification) || len(info.AccountID) == 0 {
		err = errors.New("invalid verification token")
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	accessToken, err := accountsvc.Get().VerifyOTP(c, info.AccountID, otp)
	if err != nil {
		c.JSON(StatusCodeForError(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}
