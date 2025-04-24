package error_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestContextError(t *testing.T) {
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
			name:           "Basic error",
			err:            errors.New("original error"),
			errMsg:         "error message",
			expectedError:  "error message: original error",
			expectedCode:   "CONTEXT",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Nil error",
			err:            nil,
			errMsg:         "error with nil cause",
			expectedError:  "error with nil cause: <nil>",
			expectedCode:   "CONTEXT",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Empty message",
			err:            errors.New("only error"),
			errMsg:         "",
			expectedError:  ": only error",
			expectedCode:   "CONTEXT",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			contextErr := e.NewContextError(tc.err, tc.errMsg)

			// Test Error() method
			assert.Equal(t, tc.expectedError, contextErr.Error())

			// Test ErrorObject() method
			errorObj := contextErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.expectedStatus, contextErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, contextErr.ErrorCode())
		})
	}
}
