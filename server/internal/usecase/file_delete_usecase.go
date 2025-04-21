package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
)

type FileDeleteUseCase interface {
	Execute(ctx context.Context, fileName string) e.CustomError
}

type fileDeleteUseCase struct {
	fileRepo       repository.FileRepository
	fileChunkRepo  repository.FileChunkRepository
	storageRepo    repository.StorageRepository
	baseStorageDir string
}

func NewFileDeleteUseCase(fileRepo repository.FileRepository, fileChunkRepo repository.FileChunkRepository, storageRepo repository.StorageRepository, baseStorageDir string) FileDeleteUseCase {
	return &fileDeleteUseCase{
		fileRepo:       fileRepo,
		fileChunkRepo:  fileChunkRepo,
		storageRepo:    storageRepo,
		baseStorageDir: baseStorageDir,
	}
}

// Execute deletes a file
func (uc *fileDeleteUseCase) Execute(ctx context.Context, fileName string) e.CustomError {
	// Get the file
	file, err := uc.fileRepo.GetFileByName(ctx, fileName)
	if err != nil {
		return err
	}
	if file == nil {
		return e.NewInvalidInputError(nil, fmt.Sprintf("%s not found", fileName))
	}

	// Get all file chunks
	chunks, err := uc.fileChunkRepo.GetFileChunksByParentID(ctx, file.ID)
	if err != nil {
		return err
	}

	// TODO: update chunk status to DELETE_IN_PROGRESS

	// Delete all file chunks
	for _, chunk := range chunks {
		if err := uc.storageRepo.DeleteChunk(ctx, chunk.FilePath); err != nil {
			// Log error but continue deleting other chunks
			// We want to clean up as much as possible
			fmt.Printf("failed to delete chunk %s: %v\n", chunk.FilePath, err)
		}
	}

	// Delete the file chunks from the repository
	// TODO: update chunk status to DELETED and a batch job will take care of the rest (avoid index calculation overhead)
	if err := uc.fileChunkRepo.DeleteFileChunksByParentID(ctx, file.ID); err != nil {
		return err
	}

	// Delete the file directory
	// Actual deletion is done by a batch job
	fileDirPath := filepath.Join(uc.baseStorageDir, fileName)
	if err := uc.storageRepo.DeleteDirectory(ctx, fileDirPath); err != nil {
		// Log error but continue
		fmt.Printf("failed to delete directory %s: %v\n", fileDirPath, err)
	}

	// Delete the file from the repository
	// TODO: update file status to DELETED and a batch job will take care of the rest (avoid index calculation overhead)
	if err := uc.fileRepo.DeleteFile(ctx, file.ID); err != nil {
		return err
	}

	return nil
}
