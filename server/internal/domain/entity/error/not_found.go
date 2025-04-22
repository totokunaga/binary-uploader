package error

import (
	"fmt"
	"net/http"
)

type NotFoundError struct {
	err    error
	errMsg string
}

func NewNotFoundError(err error, errMsg string) *NotFoundError {
	return &NotFoundError{
		err:    err,
		errMsg: errMsg,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: %v", e.errMsg, e.err)
}

func (e *NotFoundError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

func (e *NotFoundError) ErrorCode() string {
	return "NOT_FOUND"
}
