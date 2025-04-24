package usecase

import (
	"context"
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// DeleteUsecase handles file deletion operations
type deleteUsecase struct {
	fileClient infrastructure.FileServerHttpClient
}

// NewDeleteUsecase creates a new delete file service
func NewDeleteUsecase(fileClient infrastructure.FileServerHttpClient) *deleteUsecase {
	return &deleteUsecase{
		fileClient: fileClient,
	}
}

// Execute deletes a file on the server
func (s *deleteUsecase) Execute(ctx context.Context, targetFileName string) error {
	err := s.fileClient.DeleteFile(ctx, targetFileName)
	if err != nil {
		return fmt.Errorf("failed to delete file '%s': %w", targetFileName, err)
	}

	return nil
}
