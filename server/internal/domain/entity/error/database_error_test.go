package error_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

func TestDatabaseError(t *testing.T) {
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
			err:            errors.New("database connection failed"),
			errMsg:         "Failed to connect to database",
			expectedError:  "Failed to connect to database: database connection failed",
			expectedCode:   "DATABASE",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Nil error",
			err:            nil,
			errMsg:         "Unknown database error",
			expectedError:  "Unknown database error: <nil>",
			expectedCode:   "DATABASE",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Query error",
			err:            errors.New("invalid SQL syntax"),
			errMsg:         "Query execution failed",
			expectedError:  "Query execution failed: invalid SQL syntax",
			expectedCode:   "DATABASE",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dbErr := e.NewDatabaseError(tc.err, tc.errMsg)

			// Test Error() method
			assert.Equal(t, tc.expectedError, dbErr.Error())

			// Test ErrorObject() method
			errorObj := dbErr.ErrorObject()
			assert.NotNil(t, errorObj)

			// For nil error case, ErrorObject() formats differently than Error()
			if tc.err == nil {
				assert.Contains(t, errorObj.Error(), tc.errMsg)
			} else {
				assert.Equal(t, tc.expectedError, errorObj.Error())
			}

			// Test StatusCode() method
			assert.Equal(t, tc.expectedStatus, dbErr.StatusCode())

			// Test ErrorCode() method
			assert.Equal(t, tc.expectedCode, dbErr.ErrorCode())
		})
	}
}
