package error_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestFileStorageError(t *testing.T) {
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
			name:           "File not accessible",
			err:            errors.New("permission denied"),
			errMsg:         "Cannot access file",
			expectedError:  "Cannot access file: permission denied",
			expectedCode:   "FILE_STORAGE",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Storage full",
			err:            errors.New("not enough space"),
			errMsg:         "Storage space exceeded",
			expectedError:  "Storage space exceeded: not enough space",
			expectedCode:   "FILE_STORAGE",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Nil error",
			err:            nil,
			errMsg:         "Unknown storage error",
			expectedError:  "Unknown storage error: <nil>",
			expectedCode:   "FILE_STORAGE",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fsErr := e.NewFileStorageError(tc.err, tc.errMsg)

			// Test Error() method
			assert.Equal(t, tc.expectedError, fsErr.Error())

			// Test ErrorObject() method
			errorObj := fsErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.expectedStatus, fsErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, fsErr.ErrorCode())
		})
	}
}
