package usecase

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/database"
)

type fileGetUseCase struct {
	fileRepo database.FileRepository
}

func NewFileGetUseCase(fileRepo database.FileRepository) FileGetUseCase {
	return &fileGetUseCase{
		fileRepo: fileRepo,
	}
}

// Execute returns the list of files
func (uc *fileGetUseCase) Execute(ctx context.Context) ([]string, e.CustomError) {
	files, err := uc.fileRepo.GetFileNames(ctx)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// ExecuteGetStats returns the stats of a file
func (uc *fileGetUseCase) ExecuteGetStats(ctx context.Context, fileName string) (*entity.File, e.CustomError) {
	file, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, nil
	}

	fileStats := &entity.File{
		ID:        file.ID,
		Name:      file.Name,
		Size:      file.Size,
		Checksum:  file.Checksum,
		Status:    file.Status,
		UpdatedAt: file.UpdatedAt,
	}

	return fileStats, nil
}
