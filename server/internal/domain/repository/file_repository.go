package repository

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// FileRepository defines the interface for file data operations
type FileRepository interface {
	// CreateFile creates a new file entry in the repository
	CreateFile(ctx context.Context, file *entity.File) (uint64, e.CustomError)

	// GetFileByID retrieves a file by its ID
	GetFileByID(ctx context.Context, id uint64) (*entity.File, e.CustomError)

	// GetFileByName retrieves a file by its name
	GetFileByName(ctx context.Context, name string) (*entity.File, e.CustomError)

	// UpdateFileStatus updates the status of a file
	UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) e.CustomError

	// IncrementCompletedChunks increments the completed chunks counter of a file
	IncrementCompletedChunks(ctx context.Context, id uint64) (uint, uint, e.CustomError)

	// DeleteFile deletes a file from the repository
	DeleteFile(ctx context.Context, id uint64) e.CustomError

	// ListFiles lists all completed files
	ListFiles(ctx context.Context) ([]string, e.CustomError)
}
