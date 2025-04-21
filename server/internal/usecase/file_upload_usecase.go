package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
)

type FileUploadUseCase interface {
	ExecuteInit(ctx context.Context, fileName string, totalSize uint64, totalChunks uint) (uint64, e.CustomError)
	Execute(ctx context.Context, uploadID uint64, chunkID uint, reader io.Reader) e.CustomError
}

type fileUploadUseCase struct {
	fileRepo        repository.FileRepository
	fileChunkRepo   repository.FileChunkRepository
	storageRepo     repository.StorageRepository
	baseStorageDir  string
	uploadSizeLimit uint64
}

func NewFileUploadUseCase(
	fileRepo repository.FileRepository,
	fileChunkRepo repository.FileChunkRepository,
	storageRepo repository.StorageRepository,
	baseStorageDir string,
	uploadSizeLimit uint64,
) FileUploadUseCase {
	return &fileUploadUseCase{
		fileRepo:        fileRepo,
		fileChunkRepo:   fileChunkRepo,
		storageRepo:     storageRepo,
		baseStorageDir:  baseStorageDir,
		uploadSizeLimit: uploadSizeLimit,
	}
}

// ExecuteInit initializes a file upload and returns an upload ID
func (uc *fileUploadUseCase) ExecuteInit(ctx context.Context, fileName string, totalSize uint64, totalChunks uint) (uint64, e.CustomError) {
	// Validate the file name
	if strings.Contains(fileName, "..") {
		return 0, e.NewInvalidInputError(nil, fmt.Sprintf("invalid file name: %s", fileName))
	}

	// Check if the file already exists
	existingFile, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err != nil {
		return 0, err
	}
	if existingFile != nil {
		return 0, e.NewInvalidInputError(err, fmt.Sprintf("%s already exists", fileName))
	}

	// Check if the file size is too large // TODO: can be at the handler layer?
	if totalSize > uc.uploadSizeLimit {
		return 0, e.NewInvalidInputError(
			err,
			fmt.Sprintf("%s is too large: %d bytes. must be less than %d bytes", fileName, totalSize, uc.uploadSizeLimit),
		)
	}

	// Create a directory for the file chunks
	fileDirPath := filepath.Join(uc.baseStorageDir, fileName)
	if err := uc.storageRepo.CreateDirectory(ctx, fileDirPath); err != nil {
		return 0, err
	}

	// Create a new file entry
	file := entity.NewFile(fileName, totalSize, totalChunks)

	// Save the file to the repository
	fileID, err := uc.fileRepo.CreateFile(ctx, file)
	if err != nil {
		// Clean up the created directory
		// Not to hide the original database error and not to complicate the error handling, error occurs
		// in directory deletion is not handled here but a batch job will clean it up
		_ = uc.storageRepo.DeleteDirectory(ctx, fileDirPath)
		return 0, err
	}

	// Create file chunk entries
	if err := uc.fileChunkRepo.CreateFileChunks(ctx, fileID, totalChunks, uc.baseStorageDir, fileName); err != nil {
		// Clean up the created file and directory
		// Similarly to the previous case, error occurs in file and directory will be handled by a batch job
		_ = uc.fileRepo.DeleteFile(ctx, fileID)
		_ = uc.storageRepo.DeleteDirectory(ctx, fileDirPath)
		return 0, err
	}

	return fileID, nil
}

// Execute uploads a chunk of a file
func (uc *fileUploadUseCase) Execute(ctx context.Context, uploadID uint64, chunkID uint, reader io.Reader) e.CustomError {
	// Get the file
	file, err := uc.fileRepo.GetFileByID(ctx, uploadID)
	if err != nil || file == nil {
		return e.NewInvalidInputError(err, fmt.Sprintf("invalid upload ID: %d", uploadID))
	}

	// Check if file status is valid for uploading chunks
	// TODO: If the status is failed, remove the file chunk and create a new file chunk
	if file.Status == entity.FileStatusUploaded || file.Status == entity.FileStatusUploadFailed {
		return e.NewInvalidInputError(fmt.Errorf("file status is %s and cannot be updated", file.Status), "")
	}

	// Get the file chunk
	chunk, err := uc.fileChunkRepo.GetFileChunk(ctx, uploadID, chunkID)
	if err != nil {
		return err
	}
	if chunk == nil {
		return e.NewInvalidInputError(nil, fmt.Sprintf("invalid chunk ID %d for upload ID: %d", chunkID, uploadID))
	}

	// Check if chunk status is valid for uploading
	if chunk.Status == entity.FileChunkStatusUploaded {
		return e.NewInvalidInputError(nil, fmt.Sprintf("chunk ID: %d of upload ID: %d already uploaded", chunkID, uploadID))
	}

	// Update chunk status to processing
	if err := uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusUploadInProgress); err != nil {
		return err
	}

	// Write the chunk to storage
	if err := uc.storageRepo.WriteChunk(ctx, reader, chunk.FilePath); err != nil {
		// Update chunk status to failed
		statusUpdateErr := uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusUploadFailed)
		if statusUpdateErr != nil {
			return err
		}
		return err
	}

	// TODO: a batch job should compare the current file size and expected file size
	// Update chunk status to completed
	if err := uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusUploaded); err != nil {
		return err
	}

	// Increment completed chunks counter
	completedChunks, totalChunks, err := uc.fileRepo.IncrementCompletedChunks(ctx, uploadID)
	if err != nil {
		return err
	}

	// Check if all chunks are completed
	if completedChunks == totalChunks {
		// Update file status to completed
		if err := uc.fileRepo.UpdateFileStatus(ctx, uploadID, entity.FileStatusUploaded); err != nil {
			return err
		}
	}

	return nil
}
