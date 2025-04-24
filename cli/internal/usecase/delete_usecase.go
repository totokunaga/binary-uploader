package usecase

import (
	"context"
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// DeleteUsecase handles file deletion operations
type DeleteUsecase struct {
	fileClient infrastructure.FileServerHttpClient
}

// NewDeleteUsecase creates a new delete file service
func NewDeleteUsecase() *DeleteUsecase {
	client := infrastructure.NewFileServerV1HttpClient()
	return &DeleteUsecase{
		fileClient: client,
	}
}

// Execute deletes a file on the server
func (s *DeleteUsecase) Execute(ctx context.Context, targetFileName string) error {
	err := s.fileClient.DeleteFile(ctx, targetFileName)
	if err != nil {
		return fmt.Errorf("failed to delete file '%s': %w", targetFileName, err)
	}

	return nil
}
