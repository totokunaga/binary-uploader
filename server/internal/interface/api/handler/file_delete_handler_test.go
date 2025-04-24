package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
	"github.com/tomoya.tokunaga/server/internal/mock"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"go.uber.org/mock/gomock"
	"golang.org/x/exp/slog"
)

func setupTestFileDeleteHandler(t *testing.T, useCase usecase.FileDeleteUseCase) (*handler.FileDeleteHandler, *gin.Engine) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewFileDeleteHandler(useCase, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	return h, router
}

func TestFileDeleteHandler_Execute(t *testing.T) {
	testFileName := "test.txt"
	mockUsecaseError := e.NewMockCustomError(errors.New("usecase error"), "Usecase error", http.StatusInternalServerError)

	testCases := []struct {
		name             string
		fileNameParam    string
		setupMock        func(mockUseCase *mock.MockFileDeleteUseCase)
		expectedStatus   int
		expectedBody     *handler.FileDeleteResponse
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:          "Success",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileDeleteUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any(), testFileName).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &handler.FileDeleteResponse{Status: "OK"},
			expectError:    false,
		},
		{
			name:             "Error - Invalid file name",
			fileNameParam:    "..",
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "invalid file name",
		},
		{
			name:          "Error - Usecase Internal Error",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileDeleteUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any(), testFileName).Return(mockUsecaseError)
			},
			expectedStatus:   mockUsecaseError.StatusCode(),
			expectError:      true,
			expectedErrorMsg: mockUsecaseError.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setting up mock api server and use cases
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mock.NewMockFileDeleteUseCase(ctrl)
			if tc.setupMock != nil {
				tc.setupMock(mockUseCase)
			}

			h, router := setupTestFileDeleteHandler(t, mockUseCase)
			router.DELETE("/files/:file_name", h.Execute)

			// Executing the request
			w := httptest.NewRecorder()
			url := "/files/" + tc.fileNameParam

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, url, nil)
			router.ServeHTTP(w, req)

			// Checking the response
			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")

			if !tc.expectError {
				// checks the response body is not nil
				require.NotNil(t, tc.expectedBody, "expectedBody should not be nil for success cases")

				// checks the response body has an expected structure
				var actualBody handler.FileDeleteResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err, "Failed to unmarshal success response: %s", w.Body.String())

				// checks the response body has an expected value
				assert.Equal(t, *tc.expectedBody, actualBody)
			} else {
				var errorResponse map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)

				// checks error response has an expected structure
				require.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
				assert.Contains(t, errorResponse, "error", "Error response body does not contain 'error' field")

				// checks error message
				if tc.expectedErrorMsg != "" {
					errorMsg, ok := errorResponse["error"].(string)
					require.True(t, ok, "Error message is not a string")
					assert.Contains(t, errorMsg, tc.expectedErrorMsg, "Error message mismatch")
				}
			}
		})
	}
}
