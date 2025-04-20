package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// InitUploadService handles initializing file uploads
type InitUploadService struct {
	config               *entity.ServiceConfig
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewInitUploadService creates a new init upload service
func NewInitUploadService(config *entity.ServiceConfig) *InitUploadService {
	client := infrastructure.NewFileServerV1HttpClient(config.ServerURL)
	return &InitUploadService{
		config:               config,
		fileServerHttpClient: client,
	}
}

// Execute initializes a file upload on the server
func (s *InitUploadService) Execute(filePath string) (uploadID uint64, fileSize int64, err error) {
	// checks the existence of the file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.IsDir() {
		return 0, 0, fmt.Errorf("cannot upload a directory, please provide a file")
	}

	// checks the file isn't empty
	fileSize = fileInfo.Size() // TODO: what if the file size can't fit int64
	if fileSize == 0 {
		return 0, 0, fmt.Errorf("file is empty")
	}

	// checks the file name is valid
	fileName := filepath.Base(filePath)
	if fileName == "" {
		return 0, 0, fmt.Errorf("invalid file name")
	}

	// Calculate number of chunks (ceiling division)
	numChunks := (fileSize + s.config.ChunkSize - 1) / s.config.ChunkSize

	request := infrastructure.UploadInitRequest{
		TotalSize:   uint64(fileSize),
		TotalChunks: uint64(numChunks),
	}

	uploadID, err = s.fileServerHttpClient.InitUpload(fileName, request)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to initialize upload: %w", err)
	}

	return uploadID, fileSize, nil
}
