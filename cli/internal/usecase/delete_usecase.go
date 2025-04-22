package usecase

import (
	"fmt"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// DeleteUsecase handles file deletion operations
type DeleteUsecase struct {
	config     *entity.Config
	fileClient infrastructure.FileServerHttpClient
}

// NewDeleteUsecase creates a new delete file service
func NewDeleteUsecase(config *entity.Config) *DeleteUsecase {
	client := infrastructure.NewFileServerV1HttpClient(config.ServerURL)
	return &DeleteUsecase{
		config:     config,
		fileClient: client,
	}
}

// Execute deletes a file on the server
func (s *DeleteUsecase) Execute(targetFileName string) error {
	fmt.Printf("Initiating file deletion for '%s'...\n", targetFileName)

	err := s.fileClient.DeleteFile(targetFileName)
	if err != nil {
		return fmt.Errorf("failed to delete file '%s': %w", targetFileName, err)
	}

	return nil
}
