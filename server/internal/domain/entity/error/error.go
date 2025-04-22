package error

import (
	"fmt"
)

type CustomError interface {
	Error() string      // returns the error message
	ErrorCode() string  // returns the error code which identifies error category
	ErrorObject() error // returns the error object
	StatusCode() int    // returns the HTTP status code
}

type MockCustomError struct {
	err        error
	errMsg     string
	statusCode int
}

func NewMockCustomError(err error, errMsg string, statusCode int) *MockCustomError {
	return &MockCustomError{
		err:        err,
		errMsg:     errMsg,
		statusCode: statusCode,
	}
}

func (e *MockCustomError) Error() string {
	return fmt.Sprintf("%s: %v", e.errMsg, e.err)
}

func (e *MockCustomError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *MockCustomError) StatusCode() int {
	return e.statusCode
}

func (e *MockCustomError) ErrorCode() string {
	return "MOCK"
}
