package entity

import "time"

type FileChunkStatus string

// FileChunk represents a chunk of a file in the storage system
type FileChunk struct {
	ID          uint64     `json:"id"`
	ParentID    uint64     `json:"parent_id"`
	Status      FileStatus `json:"status"`
	ChunkNumber uint64     `json:"chunk_number"`
	Size        uint64     `json:"size"`
	FilePath    string     `json:"file_path"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewFileChunk creates a new FileChunk entity
func NewFileChunk(parentID uint64, status FileStatus, chunkNumber uint64, size uint64, filePath string) *FileChunk {
	return &FileChunk{
		ParentID:    parentID,
		Status:      status,
		ChunkNumber: chunkNumber,
		Size:        size,
		FilePath:    filePath,
	}
}
