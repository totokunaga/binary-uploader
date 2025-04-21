package usecase

import (
	"context"

	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
)

type FileListUseCase interface {
	Execute(ctx context.Context) ([]string, e.CustomError)
}

type fileListUseCase struct {
	fileRepo repository.FileRepository
}

func NewFileListUseCase(fileRepo repository.FileRepository) FileListUseCase {
	return &fileListUseCase{
		fileRepo: fileRepo,
	}
}

func (uc *fileListUseCase) Execute(ctx context.Context) ([]string, e.CustomError) {
	files, err := uc.fileRepo.ListFiles(ctx)
	if err != nil {
		return nil, err
	}

	return files, nil
}
