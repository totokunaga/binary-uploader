package infrastructure

import (
	"context"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
)

// FileServerHttpClient defines the interface for communicating with the file server
type FileServerHttpClient interface {
	// InitUpload initializes a file upload on the server
	InitUpload(ctx context.Context, fileName string, request UploadInitRequest) (*UploadInitResponse, error)

	// UploadChunk uploads a chunk to the server
	UploadChunk(ctx context.Context, uploadID uint64, chunkID int, data []byte) error

	// DeleteFile deletes a file from the server
	DeleteFile(ctx context.Context, fileName string) error

	// ListFiles lists all files on the server
	ListFiles(ctx context.Context) (*entity.ListFilesResp, error)

	// GetFileStats gets the stats of a file on the server
	GetFileStats(ctx context.Context, fileName string) (*entity.FileStatsResp, error)
}
