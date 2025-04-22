package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/mock"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestFileDeleteUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	mockStorageRepo := mock.NewMockFileStorageRepository(mockCtrl)

	baseStorageDir := "/test/storage"
	config := &entity.Config{
		BaseStorageDir: baseStorageDir,
		WorkerPoolSize: 2, // Use a small pool size for testing
	}

	uc := usecase.NewFileDeleteUseCase(config, mockFileRepo, mockStorageRepo)

	ctx := context.Background()
	testFileName := "test_file.txt"
	testFileID := uint64(1)
	testFileSize := uint64(1024)
	testChecksum := "checksum123"
	testChunks := []*entity.FileChunk{
		{ID: 1, ParentID: testFileID, ChunkNumber: 1, FilePath: "/test/storage/test_file.txt/chunk_1", Status: entity.FileStatusUploaded},
		{ID: 2, ParentID: testFileID, ChunkNumber: 2, FilePath: "/test/storage/test_file.txt/chunk_2", Status: entity.FileStatusUploaded},
	}
	testFile := &entity.File{
		ID:        testFileID,
		Name:      testFileName,
		Size:      testFileSize,
		Checksum:  testChecksum,
		Status:    entity.FileStatusUploaded,
		UpdatedAt: time.Now(),
	}
	testFileDirPath := fmt.Sprintf("%s/%s", baseStorageDir, testFileName)
	genericError := e.NewDatabaseError(errors.New("db error"), "")
	notFoundError := e.NewNotFoundError(fmt.Errorf("%s not found", testFileName), "")

	tests := []struct {
		name        string
		fileName    string
		setupMocks  func()
		expectedErr e.CustomError
	}{
		{
			name:     "Success",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(testFile, nil)
				mockFileRepo.EXPECT().GetChunksByFileID(ctx, testFileID).Return(testChunks, nil)
				// Concurrent deletion, order doesn't matter, called for each chunk
				mockStorageRepo.EXPECT().DeleteChunk(gomock.Any(), testChunks[0].FilePath).Return(nil).Times(1)
				mockStorageRepo.EXPECT().DeleteChunk(gomock.Any(), testChunks[1].FilePath).Return(nil).Times(1)
				mockStorageRepo.EXPECT().DeleteDirectory(ctx, testFileDirPath).Return(nil)
				mockFileRepo.EXPECT().DeleteFileByID(ctx, testFileID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "File Not Found",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(nil, nil)
			},
			expectedErr: notFoundError,
		},
		{
			name:     "GetFileByName Error",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(nil, genericError)
			},
			expectedErr: genericError,
		},
		{
			name:     "GetChunksByFileID Error",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(testFile, nil)
				mockFileRepo.EXPECT().GetChunksByFileID(ctx, testFileID).Return(nil, genericError)
			},
			expectedErr: genericError,
		},
		{
			name:     "DeleteDirectory Error",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(testFile, nil)
				mockFileRepo.EXPECT().GetChunksByFileID(ctx, testFileID).Return(testChunks, nil)
				mockStorageRepo.EXPECT().DeleteChunk(gomock.Any(), gomock.Any()).Return(nil).AnyTimes() // Assume chunk deletion succeeds or is ignored
				mockStorageRepo.EXPECT().DeleteDirectory(ctx, testFileDirPath).Return(genericError)
			},
			expectedErr: genericError,
		},
		{
			name:     "DeleteFileByID Error",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(testFile, nil)
				mockFileRepo.EXPECT().GetChunksByFileID(ctx, testFileID).Return(testChunks, nil)
				mockStorageRepo.EXPECT().DeleteChunk(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockStorageRepo.EXPECT().DeleteDirectory(ctx, testFileDirPath).Return(nil)
				mockFileRepo.EXPECT().DeleteFileByID(ctx, testFileID).Return(genericError)
			},
			expectedErr: genericError,
		},
		// Note: Testing DeleteChunk error is tricky due to concurrency and error suppression.
		// The current implementation logs the error but returns nil from the worker function.
		// A robust test might require checking logs or modifying the worker function for testability.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			err := uc.Execute(ctx, tc.fileName)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				// Compare specific error types or messages if needed
				assert.Equal(t, tc.expectedErr.ErrorCode(), err.ErrorCode())
				assert.Contains(t, err.Error(), tc.expectedErr.Error()) // Check if underlying error message is contained
			}
		})
	}
}
