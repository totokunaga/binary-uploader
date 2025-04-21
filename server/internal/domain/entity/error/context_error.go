package error

import (
	"fmt"
	"net/http"
)

type ContextError struct {
	err    error
	errMsg string
}

func NewContextError(err error, errMsg string) *ContextError {
	return &ContextError{
		err:    err,
		errMsg: errMsg,
	}
}

func (e *ContextError) Error() string {
	return e.errMsg
}

func (e *ContextError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *ContextError) StatusCode() int {
	return http.StatusInternalServerError
}

func (e *ContextError) ErrorCode() string {
	return "CONTEXT"
}
