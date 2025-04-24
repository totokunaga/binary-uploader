package usecase_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
	mock "github.com/tomoya.tokunaga/cli/internal/mock"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
	"github.com/tomoya.tokunaga/cli/internal/util"
	"go.uber.org/mock/gomock"
)

// Helper function to create a temporary file with content
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "testfile.txt")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create temp file")
	return filePath
}

func TestExecutePrecheck(t *testing.T) {
	ctx := context.Background()
	testContent := "hello world"
	testFilePath := createTempFile(t, testContent)
	checksum, err := util.CalculateChecksum(testFilePath)
	require.NoError(t, err)
	fileSize := int64(len(testContent))

	now := time.Now()
	serverUploadTimeoutDuration := 1 * time.Hour

	tests := []struct {
		name                 string
		mockSetup            func(mockClient *mock.MockFileServerHttpClient)
		input                *usecase.InitUploadPrecheckUsecaseInput
		wantAction           usecase.PostPrecheckAction
		wantOutput           *usecase.InitUploadPrecheckUsecaseOutput
		wantErr              bool
		expectedErrSubstring string
	}{
		{
			name: "Success: File does not exist on server",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(nil, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.ProceedWithInit,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Success: Same file exists, status Uploaded",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:   entity.FileStatusUploaded,
					Checksum: checksum,
					Size:     uint64(fileSize),
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.Exits,
			wantOutput: nil,
			wantErr:    false,
		},
		{
			name: "Success: Same file exists, status Initialized",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:   entity.FileStatusInitialized,
					Checksum: checksum,
					Size:     uint64(fileSize),
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.ProceedWithInit,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Success: Same file exists, status Failed",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:   entity.FileStatusFailed,
					Checksum: checksum,
					Size:     uint64(fileSize),
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.ProceedWithReUpload,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Success: Same file exists, status InProgress, orphaned",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:              entity.FileStatusInProgress,
					Checksum:            checksum,
					Size:                uint64(fileSize),
					UpdatedAt:           now.Add(-serverUploadTimeoutDuration).Add(-time.Second),
					UploadTimeoutSecond: serverUploadTimeoutDuration,
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.ProceedWithReUpload,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Error: Same file exists, status InProgress, not orphaned",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:              entity.FileStatusInProgress,
					Checksum:            checksum,
					Size:                uint64(fileSize),
					UpdatedAt:           now.Add(-serverUploadTimeoutDuration).Add(time.Second),
					UploadTimeoutSecond: serverUploadTimeoutDuration,
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction:           usecase.Exits,
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "remote file server processing target.txt",
		},
		{
			name: "Suggest Deletion: Different file exists (checksum mismatch)",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:   entity.FileStatusUploaded,
					Checksum: "different-checksum",
					Size:     uint64(fileSize),
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.SuggestExistingEntryDeletion,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Suggest Deletion: Different file exists (size mismatch)",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(&entity.FileStatsResp{
					Status:   entity.FileStatusUploaded,
					Checksum: checksum,
					Size:     uint64(fileSize + 1),
				}, nil)
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction: usecase.SuggestExistingEntryDeletion,
			wantOutput: &usecase.InitUploadPrecheckUsecaseOutput{
				Checksum: checksum,
				FileSize: fileSize,
			},
			wantErr: false,
		},
		{
			name: "Error: Local file does not exist",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				// No http call expected
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       "nonexistent/file.txt",
				TargetFileName: "target.txt",
			},
			wantAction:           usecase.ReturnError,
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "no such file or directory",
		},
		{
			name: "Error: GetFileStats fails",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().GetFileStats(ctx, "target.txt").Return(nil, errors.New("server error"))
			},
			input: &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       testFilePath,
				TargetFileName: "target.txt",
			},
			wantAction:           usecase.ReturnError,
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "failed to get file stats for 'target.txt': server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockFileServerHttpClient(ctrl)
			tt.mockSetup(mockClient)

			uc := usecase.NewInitUploadUsecase(mockClient)
			action, output, err := uc.ExecutePrecheck(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrSubstring != "" {
					assert.Contains(t, err.Error(), tt.expectedErrSubstring)
				}
				assert.Equal(t, tt.wantAction, action)
				assert.Nil(t, output)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAction, action)
				if tt.wantOutput != nil {
					assert.Equal(t, tt.wantOutput, output)
				} else {
					assert.Nil(t, output)
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	ctx := context.Background()
	baseChunkSize := int64(5)

	tests := []struct {
		name                  string
		mockSetup             func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64)
		inputArgs             func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput
		wantOutput            *usecase.UploadUsecaseOutput
		wantErr               bool
		expectedErrSubstring  string
		fileContent           string
		forceOriginalChecksum string
	}{
		{
			name:        "Success: Initialize new upload",
			fileContent: "hello world again",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64) {
				reqBody := infrastructure.UploadInitRequest{
					TotalSize:   fileSize,
					TotalChunks: numChunks,
					ChunkSize:   chunkSize,
					Checksum:    checksum,
					IsReUpload:  false,
				}
				mockClient.EXPECT().InitUpload(ctx, "target.txt", reqBody).Return(&infrastructure.UploadInitResponse{
					UploadID: 123,
				}, nil)
			},
			inputArgs: func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput {
				return &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   "target.txt",
					ChunkSize:        chunkSize,
					OriginalChecksum: checksum,
					IsReUpload:       false,
				}
			},
			wantOutput: &usecase.UploadUsecaseOutput{
				UploadID:              123,
				UploadChunkSize:       uint64(baseChunkSize),
				MissingChunkNumberMap: map[uint64]struct{}{},
			},
			wantErr: false,
		},
		{
			name:        "Success: Initialize re-upload with missing chunks",
			fileContent: "another file content for re-upload",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64) {
				reqBody := infrastructure.UploadInitRequest{
					TotalSize:   fileSize,
					TotalChunks: numChunks,
					ChunkSize:   chunkSize,
					Checksum:    checksum,
					IsReUpload:  true,
				}
				mockClient.EXPECT().InitUpload(ctx, "target-reup.txt", reqBody).Return(&infrastructure.UploadInitResponse{
					UploadID: 456,
					MissingChunkInfo: &entity.MissingChunkInfo{
						MaxChunkSize: 10,
						ChunkNumbers: []uint64{1, 3},
					},
				}, nil)
			},
			inputArgs: func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput {
				return &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   "target-reup.txt",
					ChunkSize:        chunkSize,
					OriginalChecksum: checksum,
					IsReUpload:       true,
				}
			},
			wantOutput: &usecase.UploadUsecaseOutput{
				UploadID:              456,
				UploadChunkSize:       10,
				MissingChunkNumberMap: map[uint64]struct{}{1: {}, 3: {}},
			},
			wantErr: false,
		},
		{
			name:        "Error: Local file does not exist",
			fileContent: "dummy content",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64) {
				// No http call expected
			},
			inputArgs: func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput {
				return &usecase.InitUploadUsecaseInput{
					FilePath:         "nonexistent/file.txt",
					TargetFileName:   "target.txt",
					ChunkSize:        chunkSize,
					OriginalChecksum: checksum,
					IsReUpload:       false,
				}
			},
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "no such file or directory",
		},
		{
			name:                  "Error: Checksum mismatch since precheck",
			fileContent:           "content for checksum test",
			forceOriginalChecksum: "different-checksum-from-precheck",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64) {
				// No http call expected
			},
			inputArgs: func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput {
				return &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   "target.txt",
					ChunkSize:        chunkSize,
					OriginalChecksum: "different-checksum-from-precheck",
					IsReUpload:       false,
				}
			},
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "file content has changed",
		},
		{
			name:        "Error: InitUpload fails on server",
			fileContent: "content for server error test",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient, fileSize int64, numChunks int64, checksum string, chunkSize int64) {
				reqBody := infrastructure.UploadInitRequest{
					TotalSize:   fileSize,
					TotalChunks: numChunks,
					ChunkSize:   chunkSize,
					Checksum:    checksum,
					IsReUpload:  false,
				}
				mockClient.EXPECT().InitUpload(ctx, "target-server-err.txt", reqBody).Return(nil, errors.New("server init error"))
			},
			inputArgs: func(filePath string, checksum string, chunkSize int64) *usecase.InitUploadUsecaseInput {
				return &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   "target-server-err.txt",
					ChunkSize:        chunkSize,
					OriginalChecksum: checksum,
					IsReUpload:       false,
				}
			},
			wantOutput:           nil,
			wantErr:              true,
			expectedErrSubstring: "failed to initialize upload for 'target-server-err.txt': server init error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			testFilePath := createTempFile(t, tt.fileContent)
			checksum, err := util.CalculateChecksum(testFilePath)
			require.NoError(t, err)
			fileSize := int64(len(tt.fileContent))
			chunkSize := baseChunkSize
			numChunks := (fileSize + chunkSize - 1) / chunkSize

			mockClient := mock.NewMockFileServerHttpClient(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient, fileSize, numChunks, checksum, chunkSize)
			}

			originalChecksumForInput := checksum
			if tt.forceOriginalChecksum != "" {
				originalChecksumForInput = tt.forceOriginalChecksum
			}

			input := tt.inputArgs(testFilePath, originalChecksumForInput, chunkSize)

			if tt.name == "Error: Local file does not exist" {
				input.FilePath = "nonexistent/file.txt"
			}

			uc := usecase.NewInitUploadUsecase(mockClient)
			output, err := uc.Execute(ctx, input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrSubstring != "" {
					assert.Contains(t, err.Error(), tt.expectedErrSubstring)
				}
				assert.Nil(t, output)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output)
			}
		})
	}
}
