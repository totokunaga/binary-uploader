package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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

// setupTestFileUploadHandler is a helper function to set up the test environment for FileUploadHandler.
func setupTestFileUploadHandler(t *testing.T, useCase usecase.FileUploadUseCase) (*handler.FileUploadHandler, *gin.Engine) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	config := &entity.Config{UploadSizeLimit: 1024 * 1024, UploadTimeoutSecond: 30}
	h := handler.NewFileUploadHandler(useCase, config, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	return h, router
}

func TestFileUploadHandler_ExecuteInit(t *testing.T) {
	testFileName := "testfile.txt"
	mockFileRecord := &entity.File{ID: 1, Name: testFileName, ChunkSize: 1024}
	mockMissingChunks := []*entity.FileChunk{{ChunkNumber: 1}, {ChunkNumber: 3}}
	mockUsecaseError := e.NewMockCustomError(errors.New("usecase init error"), "Usecase init error", http.StatusInternalServerError)

	testCases := []struct {
		name             string
		fileNameParam    string
		requestBody      handler.InitUploadRequest
		setupMock        func(mockUseCase *mock.MockFileUploadUseCase)
		expectedStatus   int
		expectedBody     *handler.InitUploadResponse
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:          "Success - No missing chunks",
			fileNameParam: testFileName,
			requestBody: handler.InitUploadRequest{
				Checksum:    "abc",
				TotalSize:   2048,
				TotalChunks: 2,
				ChunkSize:   1024,
				IsReUpload:  false,
			},
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().ExecuteInit(gomock.Any(), gomock.Any()).Return(mockFileRecord, nil, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &handler.InitUploadResponse{
				UploadID: mockFileRecord.ID,
				MissingChunkInfo: handler.MissingChunkInfo{
					MaxChunkSize: 0,          // Expect zero when no missing chunks
					ChunkNumbers: []uint64{}, // Expect empty slice
				},
			},
			expectError: false,
		},
		{
			name:          "Success - With missing chunks",
			fileNameParam: testFileName,
			requestBody: handler.InitUploadRequest{
				Checksum:    "def",
				TotalSize:   3072,
				TotalChunks: 3,
				ChunkSize:   1024,
				IsReUpload:  true,
			},
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().ExecuteInit(gomock.Any(), gomock.Any()).Return(mockFileRecord, mockMissingChunks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &handler.InitUploadResponse{
				UploadID: mockFileRecord.ID,
				MissingChunkInfo: handler.MissingChunkInfo{
					MaxChunkSize: mockFileRecord.ChunkSize,
					ChunkNumbers: []uint64{1, 3}, // Corresponding to mockMissingChunks
				},
			},
			expectError: false,
		},
		{
			name:             "Error - Missing file_name parameter",
			fileNameParam:    "", // Missing file name
			requestBody:      handler.InitUploadRequest{},
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "file_name parameter is required",
		},
		{
			name:             "Error - Invalid file_name (dot)",
			fileNameParam:    ".",
			requestBody:      handler.InitUploadRequest{},
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "invalid file name",
		},
		{
			name:             "Error - Invalid request body (missing checksum)",
			fileNameParam:    testFileName,
			requestBody:      handler.InitUploadRequest{}, // Missing required field
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "invalid request body", // Gin binding error message might vary slightly
		},
		{
			name:          "Error - File size too large",
			fileNameParam: testFileName,
			requestBody: handler.InitUploadRequest{
				Checksum:    "ghi",
				TotalSize:   2 * 1024 * 1024, // Exceeds limit defined in setup
				TotalChunks: 2,
				ChunkSize:   1024 * 1024,
			},
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "total_size is too large",
		},
		{
			name:          "Error - Usecase error",
			fileNameParam: testFileName,
			requestBody: handler.InitUploadRequest{
				Checksum:    "jkl",
				TotalSize:   1024,
				TotalChunks: 1,
				ChunkSize:   1024,
			},
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().ExecuteInit(gomock.Any(), gomock.Any()).Return(nil, nil, mockUsecaseError)
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

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mock.NewMockFileUploadUseCase(ctrl)
			if tc.setupMock != nil {
				tc.setupMock(mockUseCase)
			}

			h, router := setupTestFileUploadHandler(t, mockUseCase)
			// Define route with file_name parameter
			router.POST("/files/upload/:file_name/init", h.ExecuteInit)

			w := httptest.NewRecorder()
			url := "/files/upload/" + tc.fileNameParam + "/init"
			if tc.fileNameParam == "" {
				url = "/files/upload//init" // Simulate empty param
			}

			// Marshal request body to JSON
			var reqBodyReader io.Reader
			if tc.name != "Error - Invalid request body (missing checksum)" { // Handle invalid body case specifically
				jsonData, err := json.Marshal(tc.requestBody)
				require.NoError(t, err)
				reqBodyReader = bytes.NewBuffer(jsonData)
			} else {
				reqBodyReader = bytes.NewBuffer([]byte(`{"invalid json":`)) // Send invalid JSON
			}

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url, reqBodyReader)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")

			if !tc.expectError {
				require.NotNil(t, tc.expectedBody, "expectedBody should not be nil for success cases")
				var actualBody handler.InitUploadResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err, "Failed to unmarshal success response: %s", w.Body.String())

				// Special handling for missing chunk info slices (nil vs empty)
				if tc.expectedBody.MissingChunkInfo.ChunkNumbers == nil {
					tc.expectedBody.MissingChunkInfo.ChunkNumbers = []uint64{}
				}
				if actualBody.MissingChunkInfo.ChunkNumbers == nil {
					actualBody.MissingChunkInfo.ChunkNumbers = []uint64{}
				}
				assert.Equal(t, *tc.expectedBody, actualBody)
			} else {
				var errorResponse map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
				assert.Contains(t, errorResponse, "error", "Error response body does not contain 'error' field")
				if tc.expectedErrorMsg != "" {
					errorMsg, ok := errorResponse["error"].(string)
					require.True(t, ok, "Error message is not a string")
					assert.Contains(t, errorMsg, tc.expectedErrorMsg)
				}
			}
		})
	}
}

func TestFileUploadHandler_Execute(t *testing.T) {
	mockFileID := uint64(1)
	mockChunkNumber := uint64(1)
	mockUsecaseError := e.NewMockCustomError(errors.New("usecase execute error"), "Usecase execute error", http.StatusInternalServerError)
	mockFailRecoveryError := e.NewMockCustomError(errors.New("fail recovery error"), "Fail recovery error", http.StatusInternalServerError)

	testCases := []struct {
		name              string
		fileIDParam       string
		chunkNumberParam  string
		fileContent       string
		setupMock         func(mockUseCase *mock.MockFileUploadUseCase)
		expectedStatus    int
		expectedBody      *handler.UploadResponse
		expectError       bool
		expectedErrorMsg  string
		failRecoveryError bool // Flag to simulate error during fail recovery
	}{
		{
			name:             "Success",
			fileIDParam:      strconv.FormatUint(mockFileID, 10),
			chunkNumberParam: strconv.FormatUint(mockChunkNumber, 10),
			fileContent:      "this is chunk 1 data",
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &handler.UploadResponse{Status: "OK"},
			expectError:    false,
		},
		{
			name:             "Error - Invalid file ID",
			fileIDParam:      "invalid-id",
			chunkNumberParam: strconv.FormatUint(mockChunkNumber, 10),
			fileContent:      "some data",
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "invalid file ID: invalid-id",
		},
		{
			name:             "Error - Invalid chunk number",
			fileIDParam:      strconv.FormatUint(mockFileID, 10),
			chunkNumberParam: "invalid-chunk",
			fileContent:      "some data",
			setupMock:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "invalid chunk ID: invalid-chunk",
		},
		{
			name:             "Error - Usecase execute error (with fail recovery)",
			fileIDParam:      strconv.FormatUint(mockFileID, 10),
			chunkNumberParam: strconv.FormatUint(mockChunkNumber, 10),
			fileContent:      "some data",
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(mockUsecaseError)
				mockUseCase.EXPECT().ExecuteFailRecovery(gomock.Any(), mockFileID, mockChunkNumber).Return(nil) // Successful recovery
			},
			expectedStatus:   mockUsecaseError.StatusCode(),
			expectError:      true,
			expectedErrorMsg: mockUsecaseError.Error(),
		},
		{
			name:              "Error - Usecase execute error (with fail recovery error)",
			fileIDParam:       strconv.FormatUint(mockFileID, 10),
			chunkNumberParam:  strconv.FormatUint(mockChunkNumber, 10),
			fileContent:       "some data",
			failRecoveryError: true,
			setupMock: func(mockUseCase *mock.MockFileUploadUseCase) {
				mockUseCase.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(mockUsecaseError)
				mockUseCase.EXPECT().ExecuteFailRecovery(gomock.Any(), mockFileID, mockChunkNumber).Return(mockFailRecoveryError) // Error during recovery
			},
			expectedStatus:   mockFailRecoveryError.StatusCode(), // Expect recovery error status
			expectError:      true,
			expectedErrorMsg: mockFailRecoveryError.Error(), // Expect recovery error message
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Client disconnect tests might interfere due to goroutines, run sequentially for now.
			t.Parallel()

			// Setting up mock api server and use cases
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mock.NewMockFileUploadUseCase(ctrl)
			if tc.setupMock != nil {
				tc.setupMock(mockUseCase)
			}

			h, router := setupTestFileUploadHandler(t, mockUseCase)
			router.POST("/files/upload/:file_id/:chunk_number", h.Execute)

			w := httptest.NewRecorder()
			url := fmt.Sprintf("/files/upload/%s/%s", tc.fileIDParam, tc.chunkNumberParam)

			// Create request body with application/octet-stream
			body := bytes.NewBufferString(tc.fileContent)

			// Create context for potential cancellation
			reqCtx, cancel := context.WithCancel(context.Background())
			defer cancel() // Ensure cancel is called eventually

			req, _ := http.NewRequestWithContext(reqCtx, http.MethodPost, url, body)
			req.Header.Set("Content-Type", "application/octet-stream")

			router.ServeHTTP(w, req)

			// Checking the response
			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")

			if !tc.expectError {
				require.NotNil(t, tc.expectedBody, "expectedBody should not be nil for success cases")
				var actualBody handler.UploadResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err, "Failed to unmarshal success response: %s", w.Body.String())
				assert.Equal(t, *tc.expectedBody, actualBody)
			} else {
				// Check for standard JSON error response
				var errorResponse map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
				assert.Contains(t, errorResponse, "error", "Error response body does not contain 'error' field")
				if tc.expectedErrorMsg != "" {
					errorMsg, ok := errorResponse["error"].(string)
					require.True(t, ok, "Error message is not a string")
					assert.Contains(t, errorMsg, tc.expectedErrorMsg)
				}
			}
		})
	}
}
