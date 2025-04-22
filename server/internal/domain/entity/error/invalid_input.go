package error

import (
	"fmt"
	"net/http"
)

type InvalidInputError struct {
	err    error
	errMsg string
}

func NewInvalidInputError(err error, errMsg string) *InvalidInputError {
	return &InvalidInputError{
		err:    err,
		errMsg: errMsg,
	}
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("%s: %v", e.errMsg, e.err)
}

func (e *InvalidInputError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *InvalidInputError) StatusCode() int {
	return http.StatusBadRequest
}

func (e *InvalidInputError) ErrorCode() string {
	return "INVALID_INPUT"
}
