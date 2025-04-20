package infrastructure

// FileInfo represents a file's metadata as returned by the server
type FileInfo struct {
	Files []string `json:"files"`
}

// FileServerHttpClient defines the interface for communicating with the file server
type FileServerHttpClient interface {
	// InitUpload initializes a file upload on the server
	InitUpload(fileName string, request UploadInitRequest) (uint64, error)

	// UploadChunk uploads a chunk to the server
	UploadChunk(uploadID uint64, chunkID int, data []byte) error

	// DeleteFile deletes a file from the server
	DeleteFile(fileName string) error

	// ListFiles lists all files on the server
	ListFiles() (*FileInfo, error)
}
