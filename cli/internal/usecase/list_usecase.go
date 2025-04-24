package usecase

import (
	"context"
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// ListUsecase handles file listing operations
type ListUsecase struct {
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewListUsecase creates a new list usecase
func NewListUsecase() *ListUsecase {
	return &ListUsecase{
		fileServerHttpClient: infrastructure.NewFileServerV1HttpClient(),
	}
}

// Execute lists files available on the server
func (s *ListUsecase) Execute(ctx context.Context) (*infrastructure.FileInfo, error) {
	// Get file list from server
	fileInfos, err := s.fileServerHttpClient.ListFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return fileInfos, nil
}
