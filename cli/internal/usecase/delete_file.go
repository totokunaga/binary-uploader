package usecase

import (
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// DeleteFileService handles file deletion operations
type DeleteFileService struct {
	config     *entity.ServiceConfig
	fileClient infrastructure.FileServerHttpClient
}

// NewDeleteFileService creates a new delete file service
func NewDeleteFileService(config *entity.ServiceConfig) *DeleteFileService {
	client := infrastructure.NewFileServerV1HttpClient(config.ServerURL)
	return &DeleteFileService{
		config:     config,
		fileClient: client,
	}
}

// DeleteFile deletes a file from the server
func (s *DeleteFileService) Execute(fileName string) error {
	return s.fileClient.DeleteFile(fileName)
}
