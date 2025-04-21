package entity

import "time"

type FileChunkStatus string

const (
	FileChunkStatusUploadInitialized FileChunkStatus = "UPLOAD_INITIALIZED"
	FileChunkStatusUploadInProgress  FileChunkStatus = "UPLOAD_IN_PROGRESS"
	FileChunkStatusUploadFailed      FileChunkStatus = "UPLOAD_FAILED"
	FileChunkStatusUploaded          FileChunkStatus = "UPLOADED"
	FileChunkStatusDeleteInitialized FileChunkStatus = "DELETE_INITIALIZED"
	FileChunkStatusDeleteInProgress  FileChunkStatus = "DELETE_IN_PROGRESS"
	FileChunkStatusDeleteFailed      FileChunkStatus = "DELETE_FAILED"
	FileChunkStatusDeleted           FileChunkStatus = "DELETED"
)

// FileChunk represents a chunk of a file in the storage system
type FileChunk struct {
	ID        uint64          `json:"id"`
	ParentID  uint64          `json:"parent_id"`
	Status    FileChunkStatus `json:"status"`
	FilePath  string          `json:"file_path"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
