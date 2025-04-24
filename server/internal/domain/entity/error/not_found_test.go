package error_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		errMsg         string
		expectedError  string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "Resource not found",
			err:            fmt.Errorf("resource with ID %d not found", 123),
			errMsg:         "Resource lookup failed",
			expectedError:  "Resource lookup failed: resource with ID 123 not found",
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "File not found",
			err:            errors.New("file.txt does not exist"),
			errMsg:         "File lookup failed",
			expectedError:  "File lookup failed: file.txt does not exist",
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Nil error",
			err:            nil,
			errMsg:         "Resource not found",
			expectedError:  "Resource not found: <nil>",
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notFoundErr := e.NewNotFoundError(tc.err, tc.errMsg)

			// Test Error() method
			assert.Equal(t, tc.expectedError, notFoundErr.Error())

			// Test ErrorObject() method
			errorObj := notFoundErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.expectedStatus, notFoundErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, notFoundErr.ErrorCode())
		})
	}
}
