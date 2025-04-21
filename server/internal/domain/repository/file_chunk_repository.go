package repository

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// FileChunkRepository defines the interface for file chunk data operations
type FileChunkRepository interface {
	// CreateFileChunks creates new file chunk entries in the repository
	CreateFileChunks(ctx context.Context, fileID uint64, totalChunks uint, baseDir string, fileName string) e.CustomError

	// GetFileChunkByID retrieves a file chunk by its ID
	GetFileChunkByID(ctx context.Context, id uint64) (*entity.FileChunk, e.CustomError)

	// GetFileChunk retrieves a file chunk by parent ID and chunk ID
	GetFileChunk(ctx context.Context, parentID uint64, chunkID uint) (*entity.FileChunk, e.CustomError)

	// UpdateFileChunkStatus updates the status of a file chunk
	UpdateFileChunkStatus(ctx context.Context, id uint64, status entity.FileChunkStatus) e.CustomError

	// GetFileChunksByParentID retrieves all file chunks for a given parent ID
	GetFileChunksByParentID(ctx context.Context, parentID uint64) ([]entity.FileChunk, e.CustomError)

	// DeleteFileChunks deletes all file chunks for a given parent ID
	DeleteFileChunksByParentID(ctx context.Context, parentID uint64) e.CustomError
}
