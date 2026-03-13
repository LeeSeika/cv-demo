package api

import "github.com/leeseika/cv-demo/pkg/utils/errs"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SuccessResponse(data any) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(bizErr *errs.BizError) (int, *Response) {
	msg := bizErr.Message()
	if bizErr.Code().HTTPStatusCode() >= 500 {
		msg = "internal server error"
	}
	return bizErr.Code().HTTPStatusCode(), &Response{
		Code:    int(bizErr.Code()),
		Message: msg,
	}
}
