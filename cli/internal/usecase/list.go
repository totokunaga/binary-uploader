package usecase

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// ListService handles file listing operations
type ListService struct {
	config               *entity.ServiceConfig
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewListService creates a new list service
func NewListService(cmd *cobra.Command) *ListService {
	config := NewServiceConfig(cmd, 0, 0)
	return &ListService{
		config:               config,
		fileServerHttpClient: infrastructure.NewFileServerV1HttpClient(config.ServerURL),
	}
}

// Execute lists files available on the server
func (s *ListService) Execute() (*infrastructure.FileInfo, error) {
	// Get file list from server
	fileInfos, err := s.fileServerHttpClient.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return fileInfos, nil
}
