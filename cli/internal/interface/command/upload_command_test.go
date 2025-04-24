package command_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/interface/command"
	"github.com/tomoya.tokunaga/cli/internal/mock"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// Helper function to create a temporary file for testing uploads
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", "testupload-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, err = tempFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	return tempFile.Name()
}

func TestUploadCommandHandler_Execute(t *testing.T) {
	ctx := context.Background()
	defaultChunkSize := entity.DefaultChunkSize
	defaultChecksum := "d41d8cd98f00b204e9800998ecf8427e"
	defaultUploadID := uint64(1)

	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		fileContent   string
		userInput     string
		mockSetup     func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64)
		expectedOut   []string
		expectedError error
	}{
		{
			name:        "Success: Basic upload (ProceedWithInit)",
			args:        []string{"placeholder"},
			fileContent: "hello world",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadPrecheckUsecaseInput) (usecase.PostPrecheckAction, *usecase.InitUploadPrecheckUsecaseOutput, error) {
						assert.Equal(t, filePath, in.FilePath)
						assert.Equal(t, targetFileName, in.TargetFileName)
						return usecase.ProceedWithInit, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil
					})
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.Equal(t, filePath, in.FilePath)
						assert.Equal(t, targetFileName, in.TargetFileName)
						assert.Equal(t, checksum, in.OriginalChecksum)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						assert.False(t, in.IsReUpload)
						return &usecase.UploadUsecaseOutput{UploadID: defaultUploadID, UploadChunkSize: uint64(defaultChunkSize)}, nil
					})
				uploadMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.UploadUsecaseInput) error {
						assert.Equal(t, defaultUploadID, in.UploadID)
						assert.Equal(t, filePath, in.FilePath)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						assert.False(t, in.IsReUpload)
						assert.Nil(t, in.MissingChunkNumberMap)
						if in.ProgressCb != nil {
							in.ProgressCb(fileSize)
						}
						return nil
					})
			},
			expectedOut: []string{"Successfully uploaded!"},
		},
		{
			name:        "Success: Re-upload (ProceedWithReUpload)",
			args:        []string{"placeholder"},
			fileContent: "hello again",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				missingMap := map[uint64]struct{}{1: {}, 3: {}}
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.ProceedWithReUpload, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.True(t, in.IsReUpload)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						return &usecase.UploadUsecaseOutput{UploadID: defaultUploadID, UploadChunkSize: uint64(defaultChunkSize), MissingChunkNumberMap: missingMap}, nil
					})
				uploadMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.UploadUsecaseInput) error {
						assert.True(t, in.IsReUpload)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						assert.Equal(t, missingMap, in.MissingChunkNumberMap)
						if in.ProgressCb != nil {
							in.ProgressCb(int64(uint64(defaultChunkSize) * uint64(len(missingMap))))
						}
						return nil
					})
			},
			expectedOut: []string{"Successfully uploaded!"},
		},
		{
			name:        "Success: File already exists (Exits)",
			args:        []string{"placeholder"},
			fileContent: "identical content",
			flags:       map[string]string{"file-name": "existing.txt"},
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadPrecheckUsecaseInput) (usecase.PostPrecheckAction, *usecase.InitUploadPrecheckUsecaseOutput, error) {
						assert.Equal(t, "existing.txt", in.TargetFileName)
						return usecase.Exits, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil
					})
			},
			expectedOut: []string{"already uploaded", "Exiting..."},
		},
		{
			name:        "Success: Conflict resolved by user deletion (SuggestExistingEntryDeletion -> y)",
			args:        []string{"placeholder"},
			fileContent: "new content for conflicting file",
			flags:       map[string]string{"file-name": "conflict.txt"},
			userInput:   "y\n",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.SuggestExistingEntryDeletion, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				deleteMock.EXPECT().Execute(gomock.Any(), targetFileName).Return(nil)
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.Equal(t, filePath, in.FilePath)
						assert.Equal(t, targetFileName, in.TargetFileName)
						assert.Equal(t, checksum, in.OriginalChecksum)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						assert.False(t, in.IsReUpload)
						return &usecase.UploadUsecaseOutput{UploadID: defaultUploadID, UploadChunkSize: uint64(defaultChunkSize)}, nil
					})
				uploadMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.UploadUsecaseInput) error {
						assert.Equal(t, defaultUploadID, in.UploadID)
						assert.Equal(t, filePath, in.FilePath)
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						assert.False(t, in.IsReUpload)
						assert.Nil(t, in.MissingChunkNumberMap)
						if in.ProgressCb != nil {
							in.ProgressCb(fileSize)
						}
						return nil
					})
			},
			expectedOut: []string{
				"conflicting file",
				"delete the conflicting file",
				"Successfully deleted the conflicting file",
				"Successfully uploaded!",
			},
		},
		{
			name:        "Success: Conflict resolved by user cancel (SuggestExistingEntryDeletion -> n)",
			args:        []string{"placeholder"},
			fileContent: "different content",
			flags:       map[string]string{"file-name": "conflict_cancel.txt"},
			userInput:   "n\n",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.SuggestExistingEntryDeletion, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
			},
			expectedOut: []string{
				"conflicting file",
				"delete the conflicting file",
				"Cancelling the upload...",
			},
		},
		{
			name:        "Error: Precheck fails",
			args:        []string{"placeholder"},
			fileContent: "precheck fail content",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.PostPrecheckAction(0), nil, errors.New("precheck failed"))
			},
			expectedOut: []string{"[ERROR] Failed to initialize upload pre-check: precheck failed"},
		},
		{
			name:        "Error: Init fails",
			args:        []string{"placeholder"},
			fileContent: "init fail content",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.ProceedWithInit, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						return nil, errors.New("init failed")
					})
			},
			expectedError: errors.New("[ERROR] failed to initialize upload: init failed"),
			expectedOut:   []string{},
		},
		{
			name:        "Error: Upload fails, delete succeeds",
			args:        []string{"placeholder"},
			fileContent: "upload fail content",
			flags:       map[string]string{"file-name": "upload_fail.txt"},
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.ProceedWithInit, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						return &usecase.UploadUsecaseOutput{UploadID: defaultUploadID, UploadChunkSize: uint64(defaultChunkSize)}, nil
					})
				uploadMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.UploadUsecaseInput) error {
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						if in.ProgressCb != nil {
							in.ProgressCb(fileSize / 2)
						}
						return errors.New("chunk upload failed")
					})
				deleteMock.EXPECT().Execute(gomock.Any(), targetFileName).Return(nil)
			},
			expectedOut: []string{
				"[ERROR] Upload failed for file 'upload_fail.txt': chunk upload failed",
			},
		},
		{
			name:        "Error: Upload fails, delete fails",
			args:        []string{"placeholder"},
			fileContent: "upload fail delete fail content",
			flags:       map[string]string{"file-name": "double_fail.txt"},
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.ProceedWithInit, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				initMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.InitUploadUsecaseInput) (*usecase.UploadUsecaseOutput, error) {
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						return &usecase.UploadUsecaseOutput{UploadID: defaultUploadID, UploadChunkSize: uint64(defaultChunkSize)}, nil
					})
				uploadMock.EXPECT().Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in *usecase.UploadUsecaseInput) error {
						assert.Equal(t, int64(defaultChunkSize), in.ChunkSize)
						if in.ProgressCb != nil {
							in.ProgressCb(fileSize / 2)
						}
						return errors.New("chunk upload failed again")
					})
				deleteMock.EXPECT().Execute(gomock.Any(), targetFileName).Return(errors.New("delete also failed"))
			},
			expectedOut: []string{
				"[ERROR] Failed to delete the partially uploaded file entry: delete also failed",
				"[ERROR] Upload failed for file 'double_fail.txt': chunk upload failed again",
			},
		},
		{
			name:        "Error: Conflict resolution fails (delete error)",
			args:        []string{"placeholder"},
			fileContent: "conflict delete fail content",
			flags:       map[string]string{"file-name": "conflict_del_fail.txt"},
			userInput:   "y\n",
			mockSetup: func(initMock *mock.MockInitUploadUsecase, uploadMock *mock.MockUploadUsecase, deleteMock *mock.MockDeleteUsecase, filePath string, targetFileName string, checksum string, fileSize int64) {
				initMock.EXPECT().ExecutePrecheck(gomock.Any(), gomock.Any()).
					Return(usecase.SuggestExistingEntryDeletion, &usecase.InitUploadPrecheckUsecaseOutput{Checksum: checksum, FileSize: fileSize}, nil)
				deleteMock.EXPECT().Execute(gomock.Any(), targetFileName).Return(errors.New("delete conflict failed"))
			},
			expectedOut: []string{
				"conflicting file",
				"Attempting to delete the conflicting file",
				"[ERROR] failed to delete the conflicting file \"conflict_del_fail.txt\": delete conflict failed",
				"[ERROR] Failed to handle file conflict: failed to delete conflicting file",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInitUsecase := mock.NewMockInitUploadUsecase(ctrl)
			mockUploadUsecase := mock.NewMockUploadUsecase(ctrl)
			mockDeleteUsecase := mock.NewMockDeleteUsecase(ctrl)

			tempFilePath := createTempFile(t, tc.fileContent)
			fileBaseName := filepath.Base(tempFilePath)
			fileSize := int64(len(tc.fileContent))
			checksum := defaultChecksum

			targetFileName := fileBaseName
			if nameFlag, ok := tc.flags["file-name"]; ok {
				targetFileName = nameFlag
			}

			tc.mockSetup(mockInitUsecase, mockUploadUsecase, mockDeleteUsecase, tempFilePath, targetFileName, checksum, fileSize)

			handler := command.NewUploadCommandHandler(mockInitUsecase, mockUploadUsecase, mockDeleteUsecase)
			cmd := handler.Execute()

			var in bytes.Buffer
			if tc.userInput != "" {
				in.WriteString(tc.userInput)
				cmd.SetIn(&in)
			}

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			finalArgs := make([]string, len(tc.args))
			copy(finalArgs, tc.args)
			if len(finalArgs) > 0 {
				finalArgs[0] = tempFilePath
			}
			cmd.SetArgs(finalArgs)

			for key, val := range tc.flags {
				cmd.Flags().Set(key, val)
			}

			err := cmd.ExecuteContext(ctx)

			outputStr := out.String()
			for _, expectedSubstr := range tc.expectedOut {
				assert.Contains(t, outputStr, expectedSubstr, "Output should contain substring")
			}

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error(), "Error message mismatch")
			} else {
				assert.NoError(t, err, "Expected no error, but got one")
			}
		})
	}
}
