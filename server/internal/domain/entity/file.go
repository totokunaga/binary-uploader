package entity

import "time"

type FileStatus string

const (
	FileStatusUploadInitialized FileStatus = "UPLOAD_INITIALIZED"
	FileStatusUploadInProgress  FileStatus = "UPLOAD_IN_PROGRESS"
	FileStatusUploadFailed      FileStatus = "UPLOAD_FAILED"
	FileStatusUploaded          FileStatus = "UPLOADED"
	FileStatusDeleteInitialized FileStatus = "DELETE_INITIALIZED"
	FileStatusDeleteInProgress  FileStatus = "DELETE_IN_PROGRESS"
	FileStatusDeleteFailed      FileStatus = "DELETE_FAILED"
	FileStatusDeleted           FileStatus = "DELETED"
)

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

func NewFile(name string, size uint64, totalChunks uint) *File {
	return &File{
		Name:            name,
		Size:            size,
		Status:          FileStatusUploadInitialized,
		TotalChunks:     totalChunks,
		CompletedChunks: 0,
	}
}
