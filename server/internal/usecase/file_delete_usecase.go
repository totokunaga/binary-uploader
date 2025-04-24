package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/tomoya.tokunaga/server/internal/domain/entity" // Assuming entity package exists
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/database"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/storage"
	"github.com/tomoya.tokunaga/server/internal/util/concurrency"
)

type fileDeleteUseCase struct {
	config      *entity.Config
	fileRepo    database.FileRepository
	storageRepo storage.FileStorageRepository
}

func NewFileDeleteUseCase(config *entity.Config, fileRepo database.FileRepository, storageRepo storage.FileStorageRepository) FileDeleteUseCase {
	return &fileDeleteUseCase{
		config:      config,
		fileRepo:    fileRepo,
		storageRepo: storageRepo,
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
		return e.NewNotFoundError(fmt.Errorf("%s not found", fileName), "")
	}

	// Delete all file chunks from storage
	chunks, err := uc.fileRepo.GetChunksByFileID(ctx, file.ID)
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

	// Delete the file directory
	fileDirPath := filepath.Join(uc.config.BaseStorageDir, fileName)
	if err := uc.storageRepo.DeleteDirectory(ctx, fileDirPath); err != nil {
		return err
	}

	// Delete the file and its chunks (cascade delete)
	if err := uc.fileRepo.DeleteFileByID(ctx, file.ID); err != nil {
		return err
	}

	// Update the available space
	uc.storageRepo.UpdateAvailableSpace(int64(file.Size))

	return nil
}
