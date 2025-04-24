package error_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestMockCustomError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		errMsg        string
		statusCode    int
		expectedError string
		expectedCode  string
	}{
		{
			name:          "Standard mock error",
			err:           errors.New("underlying error"),
			errMsg:        "Test error message",
			statusCode:    http.StatusBadRequest,
			expectedError: "Test error message: underlying error",
			expectedCode:  "MOCK",
		},
		{
			name:          "Internal server error",
			err:           errors.New("server failure"),
			errMsg:        "Internal error occurred",
			statusCode:    http.StatusInternalServerError,
			expectedError: "Internal error occurred: server failure",
			expectedCode:  "MOCK",
		},
		{
			name:          "Nil error with custom status",
			err:           nil,
			errMsg:        "Custom status error",
			statusCode:    http.StatusForbidden,
			expectedError: "Custom status error: <nil>",
			expectedCode:  "MOCK",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockErr := e.NewMockCustomError(tc.err, tc.errMsg, tc.statusCode)

			// Test Error() method
			assert.Equal(t, tc.expectedError, mockErr.Error())

			// Test ErrorObject() method
			errorObj := mockErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.statusCode, mockErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, mockErr.ErrorCode())
		})
	}
}

// TestCustomErrorInterface ensures that all error types implement the CustomError interface
func TestCustomErrorInterface(t *testing.T) {
	t.Parallel()

	mockErr := e.NewMockCustomError(errors.New("test"), "mock", http.StatusBadRequest)
	contextErr := e.NewContextError(errors.New("test"), "context")
	dbErr := e.NewDatabaseError(errors.New("test"), "database")
	fsErr := e.NewFileStorageError(errors.New("test"), "storage")
	invalidErr := e.NewInvalidInputError(errors.New("test"), "invalid")
	notFoundErr := e.NewNotFoundError(errors.New("test"), "notfound")

	var customErrors []e.CustomError = []e.CustomError{
		mockErr,
		contextErr,
		dbErr,
		fsErr,
		invalidErr,
		notFoundErr,
	}

	for i, err := range customErrors {
		t.Run(fmt.Sprintf("ErrorType_%d", i), func(t *testing.T) {
			// Test Error() method
			errorStr := err.Error()
			assert.NotEmpty(t, errorStr)

			// Test ErrorCode() method
			code := err.ErrorCode()
			assert.NotEmpty(t, code)

			// Test ErrorObject() method
			obj := err.ErrorObject()
			assert.NotNil(t, obj)
			assert.Error(t, obj)

			// Test StatusCode() method
			status := err.StatusCode()
			assert.Greater(t, status, 0)
		})
	}
}
