package usecase

import (
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// ListUsecase handles file listing operations
type ListUsecase struct {
	config               *entity.Config
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewListUsecase creates a new list usecase
func NewListUsecase(config *entity.Config) *ListUsecase {
	return &ListUsecase{
		config:               config,
		fileServerHttpClient: infrastructure.NewFileServerV1HttpClient(config.ServerURL),
	}
}

// Execute lists files available on the server
func (s *ListUsecase) Execute() (*infrastructure.FileInfo, error) {
	// Get file list from server
	fileInfos, err := s.fileServerHttpClient.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return fileInfos, nil
}
