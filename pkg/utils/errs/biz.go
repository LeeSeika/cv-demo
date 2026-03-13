package errs

type BizError struct {
	code    BizCode
	message string
	err     error
}

func (e *BizError) Code() BizCode {
	return e.code
}

func (e *BizError) Message() string {
	return e.message
}

func (e *BizError) Unwrap() error {
	return e.err
}

func (e *BizError) Error() string {
	msg := e.message
	if e.err != nil {
		msg += ": " + e.err.Error()
	}
	return msg
}

func NewBizError(code BizCode, message string) *BizError {
	return &BizError{
		code:    code,
		message: message,
	}
}

func WrapBizError(code BizCode, message string, err error) *BizError {
	return &BizError{
		code:    code,
		message: message,
		err:     err,
	}
}
