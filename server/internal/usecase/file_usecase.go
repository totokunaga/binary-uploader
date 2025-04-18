package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/tomoya.tokunaga/server/internal/core/entity"
	"github.com/tomoya.tokunaga/server/internal/core/repository"
)

var (
	// ErrFileNotFound is returned when a file is not found
	ErrFileNotFound = errors.New("file not found")
	// ErrFileAlreadyExists is returned when a file already exists
	ErrFileAlreadyExists = errors.New("file already exists")
	// ErrInvalidFileName is returned when a file name is invalid
	ErrInvalidFileName = errors.New("invalid file name")
	// ErrInvalidUploadID is returned when an upload ID is invalid
	ErrInvalidUploadID = errors.New("invalid upload ID")
	// ErrInvalidChunkID is returned when a chunk ID is invalid
	ErrInvalidChunkID = errors.New("invalid chunk ID")
	// ErrChunkWriteFailed is returned when a chunk write fails
	ErrChunkWriteFailed = errors.New("chunk write failed")
	// ErrFileSizeTooLarge is returned when a file size is too large
	ErrFileSizeTooLarge = errors.New("file size too large")
)

// FileUseCase defines the interface for file use cases
type FileUseCase interface {
	// InitUpload initializes a file upload and returns an upload ID
	InitUpload(ctx context.Context, fileName string, totalSize uint64, totalChunks uint) (uint64, error)

	// UploadChunk uploads a chunk of a file
	UploadChunk(ctx context.Context, uploadID uint64, chunkID uint, reader io.Reader) error

	// DeleteFile deletes a file
	DeleteFile(ctx context.Context, fileName string) error

	// ListFiles lists all completed files
	ListFiles(ctx context.Context) ([]string, error)
}

// fileUseCase implements the FileUseCase interface
type fileUseCase struct {
	fileRepo        repository.FileRepository
	fileChunkRepo   repository.FileChunkRepository
	storageRepo     repository.StorageRepository
	baseStorageDir  string
	uploadSizeLimit uint64
}

// NewFileUseCase creates a new FileUseCase instance
func NewFileUseCase(
	fileRepo repository.FileRepository,
	fileChunkRepo repository.FileChunkRepository,
	storageRepo repository.StorageRepository,
	baseStorageDir string,
	uploadSizeLimit uint64,
) FileUseCase {
	return &fileUseCase{
		fileRepo:        fileRepo,
		fileChunkRepo:   fileChunkRepo,
		storageRepo:     storageRepo,
		baseStorageDir:  baseStorageDir,
		uploadSizeLimit: uploadSizeLimit,
	}
}

// InitUpload initializes a file upload and returns an upload ID
func (uc *fileUseCase) InitUpload(ctx context.Context, fileName string, totalSize uint64, totalChunks uint) (uint64, error) {
	// Validate the file name
	if strings.Contains(fileName, "..") {
		return 0, ErrInvalidFileName
	}

	// Check if the file already exists
	existingFile, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err == nil && existingFile != nil {
		return 0, ErrFileAlreadyExists
	}

	// Check if the file size is too large
	if totalSize > uc.uploadSizeLimit {
		return 0, ErrFileSizeTooLarge
	}

	// Create a directory for the file chunks
	fileDirPath := filepath.Join(uc.baseStorageDir, fileName)
	if err := uc.storageRepo.CreateDirectory(ctx, fileDirPath); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a new file entry
	file := &entity.File{
		Name:            fileName,
		Size:            totalSize,
		Status:          entity.FileStatusInitialized,
		TotalChunks:     totalChunks,
		CompletedChunks: 0,
	}

	// Save the file to the repository
	fileID, err := uc.fileRepo.CreateFile(ctx, file)
	if err != nil {
		// Clean up the created directory
		_ = uc.storageRepo.DeleteDirectory(ctx, fileDirPath)
		return 0, fmt.Errorf("failed to create file: %w", err)
	}

	// Create file chunk entries
	if err := uc.fileChunkRepo.CreateFileChunks(ctx, fileID, totalChunks, uc.baseStorageDir, fileName); err != nil {
		// Clean up the created file and directory
		_ = uc.fileRepo.DeleteFile(ctx, fileID)
		_ = uc.storageRepo.DeleteDirectory(ctx, fileDirPath)
		return 0, fmt.Errorf("failed to create file chunks: %w", err)
	}

	return fileID, nil
}

// UploadChunk uploads a chunk of a file
func (uc *fileUseCase) UploadChunk(ctx context.Context, uploadID uint64, chunkID uint, reader io.Reader) error {
	// Get the file
	file, err := uc.fileRepo.GetFileByID(ctx, uploadID)
	if err != nil || file == nil {
		return ErrInvalidUploadID
	}

	// Check if file status is valid for uploading chunks
	if file.Status == entity.FileStatusCompleted || file.Status == entity.FileStatusFailed {
		return fmt.Errorf("file status is %s and cannot be updated", file.Status)
	}

	// Get the file chunk
	chunk, err := uc.fileChunkRepo.GetFileChunk(ctx, uploadID, chunkID)
	if err != nil || chunk == nil {
		return ErrInvalidChunkID
	}

	// Check if chunk status is valid for uploading
	if chunk.Status == entity.FileChunkStatusCompleted {
		return nil // Chunk already uploaded, nothing to do
	}

	// Update chunk status to processing
	if err := uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusProcessing); err != nil {
		return fmt.Errorf("failed to update chunk status: %w", err)
	}

	// Write the chunk to storage
	if err := uc.storageRepo.WriteChunk(ctx, reader, chunk.FilePath); err != nil {
		// Update chunk status to failed
		_ = uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusFailed)
		return ErrChunkWriteFailed
	}

	// Update chunk status to completed
	if err := uc.fileChunkRepo.UpdateFileChunkStatus(ctx, chunk.ID, entity.FileChunkStatusCompleted); err != nil {
		return fmt.Errorf("failed to update chunk status: %w", err)
	}

	// Increment completed chunks counter
	completedChunks, totalChunks, err := uc.fileRepo.IncrementCompletedChunks(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("failed to increment completed chunks: %w", err)
	}

	// Check if all chunks are completed
	if completedChunks == totalChunks {
		// Update file status to completed
		if err := uc.fileRepo.UpdateFileStatus(ctx, uploadID, entity.FileStatusCompleted); err != nil {
			return fmt.Errorf("failed to update file status: %w", err)
		}
	}

	return nil
}

// DeleteFile deletes a file
func (uc *fileUseCase) DeleteFile(ctx context.Context, fileName string) error {
	// Get the file
	file, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err != nil || file == nil {
		return ErrFileNotFound
	}

	// Get all file chunks
	chunks, err := uc.fileChunkRepo.GetFileChunksByParentID(ctx, file.ID)
	if err != nil {
		return fmt.Errorf("failed to get file chunks: %w", err)
	}

	// Delete all file chunks
	for _, chunk := range chunks {
		if err := uc.storageRepo.DeleteChunk(ctx, chunk.FilePath); err != nil {
			// Log error but continue deleting other chunks
			// We want to clean up as much as possible
			fmt.Printf("failed to delete chunk %s: %v\n", chunk.FilePath, err)
		}
	}

	// Delete the file chunks from the repository
	if err := uc.fileChunkRepo.DeleteFileChunksByParentID(ctx, file.ID); err != nil {
		return fmt.Errorf("failed to delete file chunks from repository: %w", err)
	}

	// Delete the file directory
	fileDirPath := filepath.Join(uc.baseStorageDir, fileName)
	if err := uc.storageRepo.DeleteDirectory(ctx, fileDirPath); err != nil {
		// Log error but continue
		fmt.Printf("failed to delete directory %s: %v\n", fileDirPath, err)
	}

	// Delete the file from the repository
	if err := uc.fileRepo.DeleteFile(ctx, file.ID); err != nil {
		return fmt.Errorf("failed to delete file from repository: %w", err)
	}

	return nil
}

// ListFiles lists all completed files
func (uc *fileUseCase) ListFiles(ctx context.Context) ([]string, error) {
	return uc.fileRepo.ListFiles(ctx)
}
