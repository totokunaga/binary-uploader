package entity

import "time"

type FileStatus string

const (
	FileStatusInitialized FileStatus = "INITIALIZED"
	FileStatusInProgress  FileStatus = "IN_PROGRESS"
	FileStatusFailed      FileStatus = "FAILED"
	FileStatusUploaded    FileStatus = "UPLOADED"
)

// FileStatsResp represents the response body of the file stats request
type FileStatsResp struct {
	ID                  uint64        `json:"id"`
	Name                string        `json:"name"`
	Size                uint64        `json:"size"`
	Checksum            string        `json:"checksum"`
	Status              FileStatus    `json:"status"`
	TotalChunks         uint          `json:"total_chunks"`
	UploadedChunks      uint          `json:"uploaded_chunks"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	UploadTimeoutSecond time.Duration `json:"upload_timeout_second"`
}

type MissingChunkInfo struct {
	MaxChunkSize uint64   `json:"max_size"`
	ChunkNumbers []uint64 `json:"chunk_numbers"`
}

// ListFilesResp represents the response body of the list files request
type ListFilesResp struct {
	Files []string `json:"files"`
}
