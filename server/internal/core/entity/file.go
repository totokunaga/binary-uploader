package entity

import "time"

// FileStatus represents the status of a file upload
type FileStatus string

const (
	// FileStatusInitialized indicates the file upload has been initialized
	FileStatusInitialized FileStatus = "INITIALIZED"
	// FileStatusProcessing indicates the file is being processed
	FileStatusProcessing FileStatus = "PROCESSING"
	// FileStatusFailed indicates the file upload has failed
	FileStatusFailed FileStatus = "FAILED"
	// FileStatusCompleted indicates the file upload has completed
	FileStatusCompleted FileStatus = "COMPLETED"
)

// File represents a file in the storage system
type File struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Size            uint64     `json:"size"`
	Status          FileStatus `json:"status"`
	TotalChunks     uint       `json:"total_chunks"`
	CompletedChunks uint       `json:"completed_chunks"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	FileChunks      []FileChunk
}

// FileChunkStatus represents the status of a file chunk upload
type FileChunkStatus string

const (
	// FileChunkStatusInitialized indicates the file chunk upload has been initialized
	FileChunkStatusInitialized FileChunkStatus = "INITIALIZED"
	// FileChunkStatusProcessing indicates the file chunk is being processed
	FileChunkStatusProcessing FileChunkStatus = "PROCESSING"
	// FileChunkStatusFailed indicates the file chunk upload has failed
	FileChunkStatusFailed FileChunkStatus = "FAILED"
	// FileChunkStatusCompleted indicates the file chunk upload has completed
	FileChunkStatusCompleted FileChunkStatus = "COMPLETED"
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
