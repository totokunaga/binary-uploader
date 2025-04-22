package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/mock"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestFileGetUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	uc := usecase.NewFileGetUseCase(mockFileRepo)

	ctx := context.Background()
	genericDBError := e.NewDatabaseError(errors.New("db error"), "Failed to get file names")
	expectedFiles := []string{"file1.txt", "file2.zip"}

	tests := []struct {
		name          string
		setupMocks    func()
		expectedFiles []string
		expectedErr   e.CustomError
	}{
		{
			name: "Success",
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileNames(ctx).Return(expectedFiles, nil)
			},
			expectedFiles: expectedFiles,
			expectedErr:   nil,
		},
		{
			name: "Database Error",
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileNames(ctx).Return(nil, genericDBError)
			},
			expectedFiles: nil,
			expectedErr:   genericDBError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			files, err := uc.Execute(ctx)

			assert.Equal(t, tc.expectedFiles, files)
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.ErrorCode(), err.ErrorCode())
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestFileGetUseCase_ExecuteGetStats(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	uc := usecase.NewFileGetUseCase(mockFileRepo)

	ctx := context.Background()
	testFileName := "test.txt"
	notFoundFileName := "notfound.txt"
	errorFileName := "error.txt"
	genericDBError := e.NewDatabaseError(errors.New("db error"), "Failed to get file by name")
	testFile := &entity.File{
		ID:        1,
		Name:      testFileName,
		Size:      2048,
		Checksum:  "checksum456",
		Status:    entity.FileStatusUploaded,
		UpdatedAt: time.Now(),
	}
	// The use case only returns specific fields, so we expect a subset
	expectedFileStats := &entity.File{
		ID:        testFile.ID,
		Name:      testFile.Name,
		Size:      testFile.Size,
		Checksum:  testFile.Checksum,
		Status:    testFile.Status,
		UpdatedAt: testFile.UpdatedAt,
	}

	tests := []struct {
		name         string
		fileName     string
		setupMocks   func()
		expectedFile *entity.File
		expectedErr  e.CustomError
	}{
		{
			name:     "Success",
			fileName: testFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, testFileName).Return(testFile, nil)
			},
			expectedFile: expectedFileStats,
			expectedErr:  nil,
		},
		{
			name:     "File Not Found",
			fileName: notFoundFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, notFoundFileName).Return(nil, nil)
			},
			expectedFile: nil,
			expectedErr:  nil, // Use case returns nil, nil in this case
		},
		{
			name:     "Database Error",
			fileName: errorFileName,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileByName(ctx, errorFileName).Return(nil, genericDBError)
			},
			expectedFile: nil,
			expectedErr:  genericDBError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			fileStats, err := uc.ExecuteGetStats(ctx, tc.fileName)

			assert.Equal(t, tc.expectedFile, fileStats)
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.ErrorCode(), err.ErrorCode())
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
