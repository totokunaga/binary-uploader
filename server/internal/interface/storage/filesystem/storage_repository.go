package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tomoya.tokunaga/server/internal/core/repository"
)

// storageRepository implements the repository.StorageRepository interface
type storageRepository struct{}

// NewStorageRepository creates a new filesystem storage repository
func NewStorageRepository() repository.StorageRepository {
	return &storageRepository{}
}

// WriteChunk writes the file chunk to the storage
func (r *storageRepository) WriteChunk(ctx context.Context, reader io.Reader, filePath string) error {
	// Ensure the directory exists
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the file content in chunks of 1MB
	buffer := make([]byte, 1024*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read a chunk from the reader
			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read from reader: %w", err)
			}
			if n == 0 {
				// End of file
				return nil
			}

			// Write the chunk to the file
			if _, err := file.Write(buffer[:n]); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}
}

// DeleteChunk deletes a file chunk from the storage
func (r *storageRepository) DeleteChunk(ctx context.Context, filePath string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File already deleted or does not exist, nothing to do
		return nil
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// CreateDirectory creates a directory at the given path
func (r *storageRepository) CreateDirectory(ctx context.Context, dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// DeleteDirectory deletes a directory at the given path
func (r *storageRepository) DeleteDirectory(ctx context.Context, dirPath string) error {
	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Directory already deleted or does not exist, nothing to do
		return nil
	}

	// Delete the directory and all its contents
	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// FileExists checks if a file exists at the given path
func (r *storageRepository) FileExists(ctx context.Context, filePath string) (bool, error) {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists
		return true, nil
	}
	if os.IsNotExist(err) {
		// File does not exist
		return false, nil
	}

	// Another error occurred
	return false, fmt.Errorf("failed to check if file exists: %w", err)
}
