package database

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// FileRepository defines the interface for file data operations
type FileRepository interface {
	// GetFileByName retrieves a file by its name
	GetFileByName(ctx context.Context, name string) (*entity.File, e.CustomError)

	// GetChunksByFileID retrieves all file chunks associated with a given file ID.
	GetChunksByFileID(ctx context.Context, fileID uint64) ([]*entity.FileChunk, e.CustomError)

	// GetChunksByStatus retrieves chunks by their status
	GetChunksByStatus(ctx context.Context, fileID uint64, statuses []entity.FileStatus) ([]*entity.FileChunk, e.CustomError)

	// GetFileAndChunk retrieves a file and a specific chunk by file ID and chunk number.
	GetFileAndChunk(ctx context.Context, fileID uint64, chunkNumber uint64) (*entity.File, *entity.FileChunk, e.CustomError)

	// GetFileNames lists all completed files
	GetFileNames(ctx context.Context) ([]string, e.CustomError)

	// CreateFileWithChunks creates a file and its corresponding chunk records within a transaction.
	CreateFileWithChunks(ctx context.Context, file *entity.File, baseDir string) (*entity.File, e.CustomError)

	// DeleteFileByID deletes a file by its ID.
	DeleteFileByID(ctx context.Context, fileID uint64) e.CustomError

	// UpdateFileStatus updates the status of a file
	UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) e.CustomError

	// UpdateChunksStatus updates the status of multiple chunks identified by their IDs.
	UpdateChunksStatus(ctx context.Context, chunkIDs []uint64, status entity.FileStatus) e.CustomError

	// UpdateFileAndChunkStatus updates the status of a specific file and a specific chunk within a transaction.
	UpdateFileAndChunkStatus(ctx context.Context, fileID uint64, chunkID uint64, status entity.FileStatus) e.CustomError

	// IncrementUploadedChunks increments the uploaded chunks counter of a file
	IncrementUploadedChunks(ctx context.Context, id uint64) (uint, uint, e.CustomError)
}
