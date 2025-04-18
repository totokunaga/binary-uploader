package repository

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/core/entity"
)

// FileRepository defines the interface for file data operations
type FileRepository interface {
	// CreateFile creates a new file entry in the repository
	CreateFile(ctx context.Context, file *entity.File) (uint64, error)

	// GetFileByID retrieves a file by its ID
	GetFileByID(ctx context.Context, id uint64) (*entity.File, error)

	// GetFileByName retrieves a file by its name
	GetFileByName(ctx context.Context, name string) (*entity.File, error)

	// UpdateFileStatus updates the status of a file
	UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) error

	// IncrementCompletedChunks increments the completed chunks counter of a file
	IncrementCompletedChunks(ctx context.Context, id uint64) (uint, uint, error)

	// DeleteFile deletes a file from the repository
	DeleteFile(ctx context.Context, id uint64) error

	// ListFiles lists all completed files
	ListFiles(ctx context.Context) ([]string, error)
}
