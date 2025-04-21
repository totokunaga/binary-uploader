package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/tomoya.tokunaga/server/internal/domain/entity" // Assuming entity package exists
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
	"github.com/tomoya.tokunaga/server/internal/util/concurrency"
)

type FileDeleteUseCase interface {
	Execute(ctx context.Context, fileName string) e.CustomError
}

type fileDeleteUseCase struct {
	config        *entity.Config
	fileRepo      repository.FileRepository
	fileChunkRepo repository.FileChunkRepository
	storageRepo   repository.StorageRepository
}

func NewFileDeleteUseCase(config *entity.Config, fileRepo repository.FileRepository, fileChunkRepo repository.FileChunkRepository, storageRepo repository.StorageRepository) FileDeleteUseCase {
	return &fileDeleteUseCase{
		config:        config,
		fileRepo:      fileRepo,
		fileChunkRepo: fileChunkRepo,
		storageRepo:   storageRepo,
	}
}

// Execute deletes a file
func (uc *fileDeleteUseCase) Execute(ctx context.Context, fileName string) e.CustomError {
	// Checks if the file to delete exists
	file, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err != nil {
		return err
	}
	if file == nil {
		return e.NewInvalidInputError(nil, fmt.Sprintf("%s not found", fileName))
	}

	// Update the status of file and file chunks to DELETE_IN_PROGRESS before deleting repository record
	if err := uc.fileRepo.UpdateFileStatus(ctx, file.ID, entity.FileStatusDeleteInProgress); err != nil {
		return err
	}
	if err := uc.fileChunkRepo.UpdateFileChunkStatusByParentID(ctx, file.ID, entity.FileChunkStatusDeleteInProgress); err != nil {
		return err
	}

	// Delete all file chunks from storage
	chunks, err := uc.fileChunkRepo.GetFileChunksByParentID(ctx, file.ID)
	if err != nil {
		return err
	}

	// Use worker wp for parallel chunk deletion
	wp := concurrency.NewWorkerPool(uc.config.WorkerPoolSize)
	wp.Start(ctx)

	for _, chunk := range chunks {
		chunkPath := chunk.FilePath // use closure to capture loop variable
		wp.Submit(func() error {
			_ = uc.storageRepo.DeleteChunk(ctx, chunkPath)
			return nil
		})
	}
	wp.Wait()

	// Delete the file chunks from the repository
	// TODO: update chunk status to DELETED and a batch job will take care of the rest (avoid index calculation overhead)
	if err := uc.fileChunkRepo.DeleteFileChunksByParentID(ctx, file.ID); err != nil {
		return err
	}

	// Delete the file directory
	// TODO: Actual deletion is done by a batch job
	fileDirPath := filepath.Join(uc.config.BaseStorageDir, fileName)
	if err := uc.storageRepo.DeleteDirectory(ctx, fileDirPath); err != nil {
		return err
	}

	// Delete the file from the repository
	// TODO: update file status to DELETED and a batch job will take care of the rest (avoid index calculation overhead)
	if err := uc.fileRepo.DeleteFile(ctx, file.ID); err != nil {
		return err
	}

	return nil
}
