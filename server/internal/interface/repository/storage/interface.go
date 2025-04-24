package storage

import (
	"context"
	"io"

	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// FileStorageRepository defines the interface for file storage operations
type FileStorageRepository interface {
	// CreateDirectory creates a directory at the given path
	CreateDirectory(ctx context.Context, dirPath string) e.CustomError

	// WriteChunk writes the file chunk to the storage
	WriteChunk(ctx context.Context, reader io.Reader, filePath string) e.CustomError

	// DeleteFile deletes a file chunk from the storage
	DeleteFile(ctx context.Context, filePath string) e.CustomError

	// DeleteDirectory deletes a directory at the given path
	DeleteDirectory(ctx context.Context, dirPath string) e.CustomError

	// FileExists checks if a file exists at the given path
	FileExists(ctx context.Context, filePath string) (bool, e.CustomError)

	// GetAvailableSpace returns the available space in bytes at the given path
	GetAvailableSpace(ctx context.Context, dirPath string) uint64

	// UpdateAvailableSpace updates the available space amount
	UpdateAvailableSpace(sizeChange int64)
}
