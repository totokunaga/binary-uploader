package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/mock"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestFileUploadUseCase_ExecuteInit(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	mockStorageRepo := mock.NewMockFileStorageRepository(mockCtrl)

	baseStorageDir := "/test/uploads"
	uc := usecase.NewFileUploadUseCase(mockFileRepo, mockStorageRepo, baseStorageDir)

	ctx := context.Background()
	testFileName := "new_file.dat"
	testFileID := uint64(1)
	testTotalSize := uint64(2048)
	testChecksum := "checksum123"
	testTotalChunks := uint(2)
	testChunkSize := uint64(1024)
	fileDirPath := filepath.Join(baseStorageDir, testFileName)
	now := time.Now()

	// Available storage space
	availableSpace := uint64(10 * 1024 * 1024) // 10MB

	newFileInput := usecase.FileUploadUseCaseExecuteInitInput{
		FileName:    testFileName,
		Checksum:    testChecksum,
		TotalSize:   testTotalSize,
		TotalChunks: testTotalChunks,
		ChunkSize:   testChunkSize,
		IsReUpload:  false,
	}

	// Input with size larger than available space
	largeFileInput := usecase.FileUploadUseCaseExecuteInitInput{
		FileName:    "large_file.dat",
		Checksum:    testChecksum,
		TotalSize:   availableSpace + 1, // Larger than available space
		TotalChunks: testTotalChunks,
		ChunkSize:   testChunkSize,
		IsReUpload:  false,
	}

	newFileEntity := &entity.File{
		ID:             testFileID,
		Name:           testFileName,
		Size:           testTotalSize,
		Checksum:       testChecksum,
		ChunkSize:      testChunkSize,
		Status:         entity.FileStatusInitialized,
		TotalChunks:    testTotalChunks,
		UploadedChunks: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	existingFileInitialized := &entity.File{
		ID:        testFileID + 1,
		Name:      "existing_init.dat",
		Size:      testTotalSize,
		Checksum:  testChecksum,
		Status:    entity.FileStatusInitialized,
		UpdatedAt: now.Add(-time.Hour),
	}

	existingFileInProgress := &entity.File{
		ID:        testFileID + 2,
		Name:      "existing_inprogress.dat",
		Size:      testTotalSize,
		Checksum:  testChecksum,
		Status:    entity.FileStatusInProgress,
		UpdatedAt: now.Add(-time.Hour),
	}

	existingFileUploaded := &entity.File{
		ID:        testFileID + 3,
		Name:      "existing_uploaded.dat",
		Size:      testTotalSize,
		Checksum:  testChecksum,
		Status:    entity.FileStatusUploaded,
		UpdatedAt: now.Add(-time.Hour),
	}

	existingFileDiffChecksum := &entity.File{
		ID:        testFileID + 4,
		Name:      "existing_diff.dat",
		Size:      testTotalSize,
		Checksum:  "different_checksum",
		Status:    entity.FileStatusUploaded,
		UpdatedAt: now.Add(-time.Hour),
	}

	invalidChunks := []*entity.FileChunk{
		{ID: 10, ParentID: existingFileInProgress.ID, ChunkNumber: 1, FilePath: filepath.Join(baseStorageDir, existingFileInProgress.Name, "chunk_1"), Status: entity.FileStatusInitialized},
		{ID: 11, ParentID: existingFileInProgress.ID, ChunkNumber: 2, FilePath: filepath.Join(baseStorageDir, existingFileInProgress.Name, "chunk_2"), Status: entity.FileStatusFailed},
	}
	chunkIDsToUpdate := []uint64{invalidChunks[1].ID} // Only the one with status FAILED needs update

	dbError := e.NewDatabaseError(errors.New("db error"), "")
	storageError := e.NewFileStorageError(errors.New("storage error"), "")
	invalidInputErrorDiffContent := e.NewInvalidInputError(nil, fmt.Sprintf("%s with different content (including orphaned data) already exists", existingFileDiffChecksum.Name))
	invalidInputErrorExists := e.NewInvalidInputError(nil, fmt.Sprintf("%s with same content(including orphaned data) already exists", existingFileUploaded.Name))
	invalidInputErrorBadStatus := e.NewInvalidInputError(nil, fmt.Sprintf("existing %s is in %s status and cannot be re-uploaded", existingFileUploaded.Name, existingFileUploaded.Status))
	insufficientSpaceError := e.NewFileStorageError(
		fmt.Errorf("not enough space"),
		fmt.Sprintf("File size of %s is %d bytes, but available space is %d bytes", largeFileInput.FileName, largeFileInput.TotalSize, availableSpace),
	)

	tests := []struct {
		name                  string
		input                 usecase.FileUploadUseCaseExecuteInitInput
		setupMocks            func()
		expectedFile          *entity.File
		expectedInvalidChunks []*entity.FileChunk
		expectedErr           e.CustomError
	}{
		{
			name:  "Success - New File",
			input: newFileInput,
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, newFileInput.FileName).Return(nil, nil)
				mockFileRepo.EXPECT().CreateFileWithChunks(ctx, gomock.Any(), baseStorageDir).DoAndReturn(
					func(ctx context.Context, file *entity.File, baseDir string) (*entity.File, e.CustomError) {
						assert.Equal(t, newFileInput.FileName, file.Name)
						assert.Equal(t, newFileInput.TotalSize, file.Size)
						assert.Equal(t, newFileInput.Checksum, file.Checksum)
						assert.Equal(t, newFileInput.TotalChunks, file.TotalChunks)
						assert.Equal(t, newFileInput.ChunkSize, file.ChunkSize)
						return newFileEntity, nil
					},
				)
				mockStorageRepo.EXPECT().CreateDirectory(ctx, fileDirPath).Return(nil)
			},
			expectedFile:          newFileEntity,
			expectedInvalidChunks: nil,
			expectedErr:           nil,
		},
		{
			name:  "Error - Insufficient Storage Space",
			input: largeFileInput,
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           insufficientSpaceError,
		},
		{
			name:  "Error - New File - CreateFileWithChunks DB Error",
			input: newFileInput,
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, newFileInput.FileName).Return(nil, nil)
				mockFileRepo.EXPECT().CreateFileWithChunks(ctx, gomock.Any(), baseStorageDir).Return(nil, dbError)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           dbError,
		},
		{
			name:  "Error - New File - CreateDirectory Storage Error",
			input: newFileInput,
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, newFileInput.FileName).Return(nil, nil)
				mockFileRepo.EXPECT().CreateFileWithChunks(ctx, gomock.Any(), baseStorageDir).Return(newFileEntity, nil)
				mockStorageRepo.EXPECT().CreateDirectory(ctx, fileDirPath).Return(storageError)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           storageError,
		},
		{
			name: "Success - Existing File (INITIALIZED status)",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName: existingFileInitialized.Name, Checksum: existingFileInitialized.Checksum, TotalSize: existingFileInitialized.Size, // Match existing
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileInitialized.Name).Return(existingFileInitialized, nil)
				// Directory creation check is still performed
				mockStorageRepo.EXPECT().CreateDirectory(ctx, filepath.Join(baseStorageDir, existingFileInitialized.Name)).Return(nil)
			},
			expectedFile:          existingFileInitialized,
			expectedInvalidChunks: nil,
			expectedErr:           nil,
		},
		{
			name: "Error - Existing File (Different Content)",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName: existingFileDiffChecksum.Name, Checksum: "new_checksum", TotalSize: existingFileDiffChecksum.Size, // Different checksum
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileDiffChecksum.Name).Return(existingFileDiffChecksum, nil)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           invalidInputErrorDiffContent,
		},
		{
			name: "Error - Existing File (UPLOADED status, IsReUpload=false)",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName: existingFileUploaded.Name, Checksum: existingFileUploaded.Checksum, TotalSize: existingFileUploaded.Size, IsReUpload: false, // Match existing, no re-upload
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileUploaded.Name).Return(existingFileUploaded, nil)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           invalidInputErrorExists,
		},
		{
			name: "Success - Existing File (IN_PROGRESS status, IsReUpload=true)",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName: existingFileInProgress.Name, Checksum: existingFileInProgress.Checksum, TotalSize: existingFileInProgress.Size, IsReUpload: true, // Match existing, re-upload
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileInProgress.Name).Return(existingFileInProgress, nil)
				mockFileRepo.EXPECT().GetChunksByStatus(ctx, existingFileInProgress.ID, []entity.FileStatus{entity.FileStatusInitialized, entity.FileStatusInProgress, entity.FileStatusFailed}).Return(invalidChunks, nil)
				// Expect deletion for both invalid chunks
				mockStorageRepo.EXPECT().DeleteFile(ctx, invalidChunks[0].FilePath).Return(nil)
				mockStorageRepo.EXPECT().DeleteFile(ctx, invalidChunks[1].FilePath).Return(nil)
				// Expect status update only for the chunk that wasn't INITIALIZED, using UpdateFileAndChunkStatus
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, existingFileInProgress.ID, chunkIDsToUpdate, entity.FileStatusInitialized).Return(nil)
				// Expect UpdateAvailableSpace to be called
				mockStorageRepo.EXPECT().UpdateAvailableSpace(int64(existingFileInProgress.Size))
			},
			expectedFile:          existingFileInProgress,
			expectedInvalidChunks: invalidChunks,
			expectedErr:           nil,
		},
		{
			name: "Error - Existing File (UPLOADED status, IsReUpload=true)",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName: existingFileUploaded.Name, Checksum: existingFileUploaded.Checksum, TotalSize: existingFileUploaded.Size, IsReUpload: true, // Match existing, re-upload
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileUploaded.Name).Return(existingFileUploaded, nil)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           invalidInputErrorBadStatus,
		},
		{
			name: "Error - ReUpload - GetChunksByStatus DB Error",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName:   existingFileInProgress.Name,
				TotalSize:  existingFileInProgress.Size,
				Checksum:   existingFileInProgress.Checksum,
				IsReUpload: true,
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileInProgress.Name).Return(existingFileInProgress, nil)
				mockFileRepo.EXPECT().GetChunksByStatus(ctx, existingFileInProgress.ID, gomock.Any()).Return(nil, dbError)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           dbError,
		},
		{
			name: "Error - ReUpload - DeleteFile Storage Error",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName:   existingFileInProgress.Name,
				TotalSize:  existingFileInProgress.Size,
				Checksum:   existingFileInProgress.Checksum,
				IsReUpload: true,
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileInProgress.Name).Return(existingFileInProgress, nil)
				mockFileRepo.EXPECT().GetChunksByStatus(ctx, existingFileInProgress.ID, gomock.Any()).Return(invalidChunks, nil)
				mockStorageRepo.EXPECT().DeleteFile(ctx, invalidChunks[0].FilePath).Return(storageError)
				// No need to mock the second DeleteFile or UpdateChunksStatus as it errors out
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           storageError,
		},
		{
			name: "Error - ReUpload - UpdateChunksStatus DB Error",
			input: usecase.FileUploadUseCaseExecuteInitInput{
				FileName:   existingFileInProgress.Name,
				TotalSize:  existingFileInProgress.Size,
				Checksum:   existingFileInProgress.Checksum,
				IsReUpload: true,
			},
			setupMocks: func() {
				mockStorageRepo.EXPECT().GetAvailableSpace(ctx, baseStorageDir).Return(availableSpace)
				mockFileRepo.EXPECT().GetFileByName(ctx, existingFileInProgress.Name).Return(existingFileInProgress, nil)
				mockFileRepo.EXPECT().GetChunksByStatus(ctx, existingFileInProgress.ID, gomock.Any()).Return(invalidChunks, nil)
				mockStorageRepo.EXPECT().DeleteFile(ctx, gomock.Any()).Return(nil).Times(len(invalidChunks))
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, existingFileInProgress.ID, chunkIDsToUpdate, entity.FileStatusInitialized).Return(dbError)
			},
			expectedFile:          nil,
			expectedInvalidChunks: nil,
			expectedErr:           dbError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			file, invalid, err := uc.ExecuteInit(ctx, tc.input)

			assert.Equal(t, tc.expectedFile, file)
			assert.Equal(t, tc.expectedInvalidChunks, invalid)

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

func TestFileUploadUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	mockStorageRepo := mock.NewMockFileStorageRepository(mockCtrl)

	baseStorageDir := "/test/uploads"
	uc := usecase.NewFileUploadUseCase(mockFileRepo, mockStorageRepo, baseStorageDir)

	ctx := context.Background()
	testFileID := uint64(50)
	testChunkNumber := uint64(1)
	testTotalChunks := uint(2)
	testChunkID := uint64(100)
	testChunkPath := filepath.Join(baseStorageDir, "uploading_file.dat", fmt.Sprintf("chunk_%d", testChunkNumber))
	testReader := bytes.NewReader([]byte("chunk data"))

	fileInitialized := &entity.File{ID: testFileID, Status: entity.FileStatusInitialized, TotalChunks: testTotalChunks}
	fileInProgress := &entity.File{ID: testFileID, Status: entity.FileStatusInProgress, TotalChunks: testTotalChunks}
	fileUploaded := &entity.File{ID: testFileID, Status: entity.FileStatusUploaded, TotalChunks: testTotalChunks}

	chunkInitialized := &entity.FileChunk{ID: testChunkID, ParentID: testFileID, ChunkNumber: testChunkNumber, FilePath: testChunkPath, Status: entity.FileStatusInitialized}
	chunkUploaded := &entity.FileChunk{ID: testChunkID, ParentID: testFileID, ChunkNumber: testChunkNumber, FilePath: testChunkPath, Status: entity.FileStatusUploaded}

	dbError := e.NewDatabaseError(errors.New("db error"), "")
	storageError := e.NewFileStorageError(errors.New("storage error"), "")
	invalidInputNilError := e.NewInvalidInputError(nil, fmt.Sprintf("data not found for (file ID, chunk ID) = (%d, %d)", testFileID, testChunkNumber))
	invalidInputStatusError := e.NewInvalidInputError(fmt.Errorf("file upload needs to be initialized"), "")

	input := usecase.FileUploadUseCaseExecuteInput{
		FileID:      testFileID,
		ChunkNumber: testChunkNumber,
		Reader:      testReader,
	}

	tests := []struct {
		name        string
		input       usecase.FileUploadUseCaseExecuteInput
		setupMocks  func()
		expectedErr e.CustomError
	}{
		{
			name:  "Success - First Chunk (File INITIALIZED)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInitialized, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, testFileID, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(nil)
				mockFileRepo.EXPECT().CountChunksByStatus(ctx, testFileID, entity.FileStatusUploaded).Return(int64(1), int64(testTotalChunks), nil)
			},
			expectedErr: nil,
		},
		{
			name:  "Success - Subsequent Chunk (File IN_PROGRESS)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(nil)
				mockFileRepo.EXPECT().CountChunksByStatus(ctx, testFileID, entity.FileStatusUploaded).Return(int64(1), int64(testTotalChunks), nil)
			},
			expectedErr: nil,
		},
		{
			name:  "Success - Last Chunk",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(nil)
				mockFileRepo.EXPECT().CountChunksByStatus(ctx, testFileID, entity.FileStatusUploaded).Return(int64(testTotalChunks), int64(testTotalChunks), nil)
				mockFileRepo.EXPECT().UpdateFileStatus(ctx, testFileID, entity.FileStatusUploaded).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:  "Error - GetFileAndChunk Not Found (nil, nil)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(nil, nil, nil)
			},
			expectedErr: invalidInputNilError,
		},
		{
			name:  "Error - GetFileAndChunk DB Error",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(nil, nil, dbError)
			},
			expectedErr: dbError,
		},
		{
			name:  "Error - Invalid File Status (UPLOADED)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileUploaded, chunkInitialized, nil)
			},
			expectedErr: invalidInputStatusError,
		},
		{
			name:  "Error - Invalid Chunk Status (UPLOADED)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkUploaded, nil)
			},
			expectedErr: invalidInputStatusError,
		},
		{
			name:  "Error - UpdateFileAndChunkStatus DB Error (First Chunk)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInitialized, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, testFileID, []uint64{testChunkID}, entity.FileStatusInProgress).Return(dbError)
			},
			expectedErr: dbError,
		},
		{
			name:  "Error - UpdateChunksStatus DB Error (Subsequent Chunk, to IN_PROGRESS)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(dbError)
			},
			expectedErr: dbError,
		},
		{
			name:  "Error - WriteChunk Storage Error",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(storageError)
			},
			expectedErr: storageError,
		},
		{
			name:  "Error - UpdateChunksStatus DB Error (to UPLOADED)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(dbError)
			},
			expectedErr: dbError,
		},
		{
			name:  "Error - CountChunksByStatus DB Error",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(nil)
				mockFileRepo.EXPECT().CountChunksByStatus(ctx, testFileID, entity.FileStatusUploaded).Return(int64(0), int64(0), dbError)
			},
			expectedErr: dbError,
		},
		{
			name:  "Error - UpdateFileStatus DB Error (Last Chunk)",
			input: input,
			setupMocks: func() {
				mockFileRepo.EXPECT().GetFileAndChunk(ctx, testFileID, testChunkNumber).Return(fileInProgress, chunkInitialized, nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusInProgress).Return(nil)
				mockStorageRepo.EXPECT().WriteChunk(ctx, input.Reader, testChunkPath).Return(nil)
				mockFileRepo.EXPECT().UpdateChunksStatus(ctx, []uint64{testChunkID}, entity.FileStatusUploaded).Return(nil)
				mockFileRepo.EXPECT().CountChunksByStatus(ctx, testFileID, entity.FileStatusUploaded).Return(int64(testTotalChunks), int64(testTotalChunks), nil)
				mockFileRepo.EXPECT().UpdateFileStatus(ctx, testFileID, entity.FileStatusUploaded).Return(dbError)
			},
			expectedErr: dbError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset reader for each test if it's the same instance
			if seeker, ok := tc.input.Reader.(io.Seeker); ok {
				_, _ = seeker.Seek(0, io.SeekStart)
			}

			tc.setupMocks()
			err := uc.Execute(ctx, tc.input)

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

func TestFileUploadUseCase_ExecuteFailRecovery(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileRepo := mock.NewMockFileRepository(mockCtrl)
	mockStorageRepo := mock.NewMockFileStorageRepository(mockCtrl) // Not used but needed for constructor

	baseStorageDir := "/test/uploads"
	uc := usecase.NewFileUploadUseCase(mockFileRepo, mockStorageRepo, baseStorageDir)

	ctx := context.Background()
	testFileID := uint64(70)
	testChunkID := uint64(170)
	dbError := e.NewDatabaseError(errors.New("db error"), "")

	tests := []struct {
		name        string
		fileID      uint64
		chunkID     uint64
		setupMocks  func()
		expectedErr e.CustomError
	}{
		{
			name:    "Success",
			fileID:  testFileID,
			chunkID: testChunkID,
			setupMocks: func() {
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, testFileID, []uint64{testChunkID}, entity.FileStatusFailed).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Error - DB Error",
			fileID:  testFileID,
			chunkID: testChunkID,
			setupMocks: func() {
				mockFileRepo.EXPECT().UpdateFileAndChunkStatus(ctx, testFileID, []uint64{testChunkID}, entity.FileStatusFailed).Return(dbError)
			},
			expectedErr: dbError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			err := uc.ExecuteFailRecovery(ctx, tc.fileID, tc.chunkID)

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
