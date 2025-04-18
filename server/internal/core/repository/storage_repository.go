package repository

import (
	"context"
	"io"
)

// StorageRepository defines the interface for file storage operations
type StorageRepository interface {
	// WriteChunk writes the file chunk to the storage
	WriteChunk(ctx context.Context, reader io.Reader, filePath string) error

	// DeleteChunk deletes a file chunk from the storage
	DeleteChunk(ctx context.Context, filePath string) error

	// CreateDirectory creates a directory at the given path
	CreateDirectory(ctx context.Context, dirPath string) error

	// DeleteDirectory deletes a directory at the given path
	DeleteDirectory(ctx context.Context, dirPath string) error

	// FileExists checks if a file exists at the given path
	FileExists(ctx context.Context, filePath string) (bool, error)
}
