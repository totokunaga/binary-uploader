package error

import (
	"fmt"
	"net/http"
)

type DatabaseError struct {
	err    error
	errMsg string
}

func NewDatabaseError(err error, errMsg string) *DatabaseError {
	return &DatabaseError{
		err:    err,
		errMsg: errMsg,
	}
}

func (e *DatabaseError) Error() string {
	return e.errMsg
}

func (e *DatabaseError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *DatabaseError) StatusCode() int {
	return http.StatusInternalServerError
}

func (e *DatabaseError) ErrorCode() string {
	return "DATABASE"
}
