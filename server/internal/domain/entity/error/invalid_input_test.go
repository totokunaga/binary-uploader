package error_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestInvalidInputError(t *testing.T) {
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
			name:           "Missing required field",
			err:            errors.New("field 'name' is required"),
			errMsg:         "Validation failed",
			expectedError:  "Validation failed: field 'name' is required",
			expectedCode:   "INVALID_INPUT",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid format",
			err:            errors.New("invalid email format"),
			errMsg:         "Email validation failed",
			expectedError:  "Email validation failed: invalid email format",
			expectedCode:   "INVALID_INPUT",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nil error",
			err:            nil,
			errMsg:         "Invalid input provided",
			expectedError:  "Invalid input provided: <nil>",
			expectedCode:   "INVALID_INPUT",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inputErr := e.NewInvalidInputError(tc.err, tc.errMsg)

			// Test Error() method
			assert.Equal(t, tc.expectedError, inputErr.Error())

			// Test ErrorObject() method
			errorObj := inputErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.expectedStatus, inputErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, inputErr.ErrorCode())
		})
	}
}
