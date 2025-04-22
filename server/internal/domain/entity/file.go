package entity

import "time"

type FileStatus string

const (
	FileStatusInitialized FileStatus = "INITIALIZED"
	FileStatusInProgress  FileStatus = "IN_PROGRESS"
	FileStatusFailed      FileStatus = "FAILED"
	FileStatusUploaded    FileStatus = "UPLOADED"
)

type File struct {
	ID             uint64     `json:"id"`
	Name           string     `json:"name"`
	Size           uint64     `json:"size"`
	Checksum       string     `json:"checksum"`
	ChunkSize      uint64     `json:"chunk_size"`
	Status         FileStatus `json:"status"`
	TotalChunks    uint       `json:"total_chunks"`
	UploadedChunks uint       `json:"uploaded_chunks"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FileChunks     []FileChunk
}

func NewFile(name string, size uint64, checksum string, totalChunks uint, chunkSize uint64) *File {
	return &File{
		Name:           name,
		Size:           size,
		Checksum:       checksum,
		ChunkSize:      chunkSize,
		Status:         FileStatusInitialized,
		TotalChunks:    totalChunks,
		UploadedChunks: 0,
	}
}
