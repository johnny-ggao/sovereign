package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestAppErrorMessage(t *testing.T) {
	err := New(400, "BAD", "bad request")
	if err.Error() != "bad request" {
		t.Errorf("Error() = %q, want %q", err.Error(), "bad request")
	}
}

func TestAppErrorWrap(t *testing.T) {
	inner := fmt.Errorf("db timeout")
	wrapped := Wrap(ErrInternal, inner)

	if wrapped.HTTPStatus != 500 {
		t.Errorf("HTTPStatus = %d, want 500", wrapped.HTTPStatus)
	}
	if wrapped.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q, want INTERNAL_ERROR", wrapped.Code)
	}
	if !errors.Is(wrapped, inner) {
		t.Error("Unwrap should return inner error")
	}
	// Original sentinel should not be mutated
	if ErrInternal.Internal != nil {
		t.Error("sentinel error should not be mutated")
	}
}

func TestErrorsAs(t *testing.T) {
	err := Wrap(ErrInvalidCredentials, fmt.Errorf("wrong hash"))

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("errors.As should match AppError")
	}
	if appErr.Code != "AUTH_INVALID_CREDENTIALS" {
		t.Errorf("Code = %q, want AUTH_INVALID_CREDENTIALS", appErr.Code)
	}
}
