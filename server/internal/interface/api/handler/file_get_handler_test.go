package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
	"github.com/tomoya.tokunaga/server/internal/mock"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"go.uber.org/mock/gomock"
	"golang.org/x/exp/slog"
)

func setupTestFileGetHandler(t *testing.T, useCase usecase.FileGetUseCase) (*handler.FileGetHandler, *gin.Engine) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	config := &entity.Config{UploadTimeoutSecond: 30}
	h := handler.NewFileGetHandler(useCase, config, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	return h, router
}

func TestFileGetHandler_Execute(t *testing.T) {
	mockUsecaseError := e.NewMockCustomError(errors.New("usecase error"), "Usecase error", http.StatusInternalServerError)

	testCases := []struct {
		name             string
		setupMock        func(mockUseCase *mock.MockFileGetUseCase)
		expectedStatus   int
		expectedBody     *handler.FileGetResponse
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "Success",
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any()).Return([]string{"file1.txt", "file2.jpg"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &handler.FileGetResponse{Files: []string{"file1.txt", "file2.jpg"}},
			expectError:    false,
		},
		{
			name: "Error - Usecase Internal Error",
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any()).Return(nil, mockUsecaseError)
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

			mockUseCase := mock.NewMockFileGetUseCase(ctrl)
			if tc.setupMock != nil {
				tc.setupMock(mockUseCase)
			}

			h, router := setupTestFileGetHandler(t, mockUseCase)
			router.GET("/files", h.Execute)

			// Executing the request
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/files", nil)
			router.ServeHTTP(w, req)

			// Checking the response
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectError {
				// checks the response body is not nil
				require.NotNil(t, tc.expectedBody, "expectedBody should not be nil for success cases")

				// checks the response body has an expected structure
				var actualBody handler.FileGetResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err)

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
					assert.Contains(t, errorMsg, tc.expectedErrorMsg)
				}
			}
		})
	}
}

func TestFileGetHandler_ExecuteGetStats(t *testing.T) {
	testFileName := "testfile.txt"
	testFile := &entity.File{
		Name:        testFileName,
		Size:        1024,
		Status:      entity.FileStatusUploaded,
		Checksum:    "dummychecksum",
		ChunkSize:   512,
		TotalChunks: 2,
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	uploadTimeoutDuration := 30 * time.Second

	mockNotFoundError := e.NewNotFoundError(errors.New("file not found"), "File stats not found for "+testFileName)
	mockForbiddenError := e.NewInvalidInputError(errors.New("access denied"), "User does not have permission to access file stats")
	mockUsecaseError := e.NewMockCustomError(errors.New("usecase error"), "Usecase error", http.StatusInternalServerError)

	testCases := []struct {
		name             string
		fileNameParam    string
		setupMock        func(mockUseCase *mock.MockFileGetUseCase)
		expectedStatus   int
		expectedBody     *handler.FileGetStatsResponse
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:          "Success",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().ExecuteGetStats(gomock.Any(), testFileName).Return(testFile, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &handler.FileGetStatsResponse{
				File:                testFile,
				UploadTimeoutSecond: uploadTimeoutDuration,
			},
			expectError: false,
		},
		{
			name:             "Error - Missing file_name parameter",
			fileNameParam:    "",
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "file_name parameter is required",
		},
		{
			name:          "Error - General usecase error",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().ExecuteGetStats(gomock.Any(), testFileName).Return(nil, mockUsecaseError)
			},
			expectedStatus:   mockUsecaseError.StatusCode(),
			expectError:      true,
			expectedErrorMsg: mockUsecaseError.Error(),
		},
		{
			name:          "Error - File not found (nil result from use case)",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().ExecuteGetStats(gomock.Any(), testFileName).Return(nil, nil)
			},
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "file not found",
		},
		{
			name:          "Error - File not found (NotFoundError from use case)",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().ExecuteGetStats(gomock.Any(), testFileName).Return(nil, mockNotFoundError)
			},
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: mockNotFoundError.Error(),
		},
		{
			name:          "Error - Specific domain error from use case (Simulated Forbidden)",
			fileNameParam: testFileName,
			setupMock: func(mockUseCase *mock.MockFileGetUseCase) {
				mockUseCase.EXPECT().ExecuteGetStats(gomock.Any(), testFileName).Return(nil, mockForbiddenError)
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: mockForbiddenError.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setting up mock api server and use cases
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mock.NewMockFileGetUseCase(ctrl)
			if tc.setupMock != nil {
				tc.setupMock(mockUseCase)
			}

			h, router := setupTestFileGetHandler(t, mockUseCase)
			router.GET("/files/:file_name/stats", h.ExecuteGetStats)

			// Executing the request
			w := httptest.NewRecorder()
			url := "/files/" + tc.fileNameParam + "/stats"
			if tc.fileNameParam == "" {
				url = "/files//stats"
			}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			router.ServeHTTP(w, req)

			// Checking the response
			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")

			if !tc.expectError {
				// checks the response body is not nil
				require.NotNil(t, tc.expectedBody, "expectedBody should not be nil for success cases")

				// checks the response body has an expected structure
				var actualBody handler.FileGetStatsResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err, "Failed to unmarshal success response: %s", w.Body.String())

				// checks the response body has an expected value
				require.NotNil(t, actualBody.File, "Actual body file is nil")
				require.NotNil(t, tc.expectedBody.File, "Expected body file is nil")
				assert.Equal(t, tc.expectedBody.Name, actualBody.Name)
				assert.Equal(t, tc.expectedBody.Size, actualBody.Size)
				assert.Equal(t, tc.expectedBody.Status, actualBody.Status)
				assert.Equal(t, tc.expectedBody.Checksum, actualBody.Checksum)
				assert.Equal(t, tc.expectedBody.ChunkSize, actualBody.ChunkSize)
				assert.Equal(t, tc.expectedBody.TotalChunks, actualBody.TotalChunks)
				assert.WithinDuration(t, tc.expectedBody.CreatedAt, actualBody.CreatedAt, time.Second)
			} else {
				var errorResponse map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)

				// checks error response has an expected structure
				require.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
				assert.Contains(t, errorResponse, "error", "Error response body does not contain 'error' field")

				// checks error message
				errorMsg, ok := errorResponse["error"].(string)
				require.True(t, ok, "Error message is not a string")

				if tc.expectedErrorMsg != "" {
					assert.Contains(t, errorMsg, tc.expectedErrorMsg, "Error message mismatch")
				}
			}
		})
	}
}
