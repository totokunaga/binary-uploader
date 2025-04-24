package usecase

import (
	"context"
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// ListUsecase handles file listing operations
type listUsecase struct {
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewListUsecase creates a new list usecase
func NewListUsecase(fileClient infrastructure.FileServerHttpClient) *listUsecase {
	return &listUsecase{
		fileServerHttpClient: fileClient,
	}
}

// Execute lists files available on the server
func (s *listUsecase) Execute(ctx context.Context) (*entity.ListFilesResp, error) {
	// Get file list from server
	fileInfos, err := s.fileServerHttpClient.ListFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return fileInfos, nil
}
