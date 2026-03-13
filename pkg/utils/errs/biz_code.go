package errs

import "net/http"

type BizCode uint

const (
	ErrBadRequest       BizCode = 1000 // HTTP 400
	ErrUnauthenticated  BizCode = 1100 // HTTP 401
	ErrResourceNotFound BizCode = 1201 // HTTP 404
	ErrResourceExists   BizCode = 1401 // HTTP 409
	ErrInternalServer   BizCode = 1700 // HTTP 500
)

func (c BizCode) HTTPStatusCode() int {
	switch c {
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrUnauthenticated:
		return http.StatusUnauthorized
	case ErrResourceNotFound:
		return http.StatusNotFound
	case ErrResourceExists:
		return http.StatusConflict
	case ErrInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
