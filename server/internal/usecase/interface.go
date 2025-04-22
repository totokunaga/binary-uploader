package usecase

import (
	"context"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

type FileGetUseCase interface {
	Execute(ctx context.Context) ([]string, e.CustomError)
	ExecuteGetStats(ctx context.Context, fileName string) (*entity.File, e.CustomError)
}

type FileUploadUseCase interface {
	ExecuteInit(ctx context.Context, input FileUploadUseCaseExecuteInitInput) (*entity.File, []*entity.FileChunk, e.CustomError)
	Execute(ctx context.Context, input FileUploadUseCaseExecuteInput) e.CustomError
	ExecuteFailRecovery(ctx context.Context, fileID uint64, chunkID uint64) e.CustomError
}

type FileDeleteUseCase interface {
	Execute(ctx context.Context, fileName string) e.CustomError
}
