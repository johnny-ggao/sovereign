package cobo

import "fmt"

type APIError struct {
	HTTPStatus int    `json:"-"`
	ErrorCode  int    `json:"error_code"`
	ErrorMsg   string `json:"error_message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cobo api error %d: %s (http %d)", e.ErrorCode, e.ErrorMsg, e.HTTPStatus)
}

func isRetryable(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.HTTPStatus >= 500
	}
	return false
}
