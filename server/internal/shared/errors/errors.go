package apperr

import "fmt"

type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
	Internal   error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

func Wrap(base *AppError, internal error) *AppError {
	return &AppError{
		HTTPStatus: base.HTTPStatus,
		Code:       base.Code,
		Message:    base.Message,
		Internal:   internal,
	}
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}
