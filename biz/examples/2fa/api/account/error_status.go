package account

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

func StatusCodeForError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))

	switch {
	case strings.Contains(message, "required"):
		return http.StatusBadRequest
	case strings.Contains(message, "invalid request"):
		return http.StatusBadRequest
	case strings.Contains(message, "already registered"):
		return http.StatusConflict
	case strings.Contains(message, "invalid"):
		return http.StatusUnauthorized
	case strings.Contains(message, "incorrect"):
		return http.StatusUnauthorized
	case strings.Contains(message, "token"):
		return http.StatusUnauthorized
	case strings.Contains(message, "unauthorized"):
		return http.StatusUnauthorized
	case strings.Contains(message, "not found"):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
