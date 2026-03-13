package api

import (
	"errors"
	"net/http"

	"github.com/leeseika/cv-demo/pkg/utils/errs"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SuccessResponse(message string, data any) *Response {
	return &Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

func BadRequestResponse(message string) *Response {
	return &Response{
		Code:    int(errs.ErrBadRequest),
		Message: message,
	}
}

func BizErrorResponse(err error) (int, *Response) {
	var bizErr *errs.BizError

	switch {
	case errors.As(err, &bizErr):
		msg := bizErr.Message()
		if bizErr.Code().HTTPStatusCode() >= 500 {
			msg = "internal server error"
		}
		return bizErr.Code().HTTPStatusCode(), &Response{
			Code:    int(bizErr.Code()),
			Message: msg,
		}
	default:
		return http.StatusInternalServerError, &Response{
			Code:    int(errs.ErrInternalServer),
			Message: "internal server error",
		}
	}
}
