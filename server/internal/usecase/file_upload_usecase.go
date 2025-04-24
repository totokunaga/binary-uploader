package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/database"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/storage"
)

type FileUploadUseCaseExecuteInitInput struct {
	FileName    string
	Checksum    string
	TotalSize   uint64
	TotalChunks uint
	ChunkSize   uint64
	IsReUpload  bool
}

type FileUploadUseCaseExecuteInput struct {
	FileID      uint64
	ChunkNumber uint64
	Reader      io.Reader
}

type fileUploadUseCase struct {
	fileRepo       database.FileRepository
	storageRepo    storage.FileStorageRepository
	baseStorageDir string
}

func NewFileUploadUseCase(
	fileRepo database.FileRepository,
	storageRepo storage.FileStorageRepository,
	baseStorageDir string,
) FileUploadUseCase {
	return &fileUploadUseCase{
		fileRepo:       fileRepo,
		storageRepo:    storageRepo,
		baseStorageDir: baseStorageDir,
	}
}

// ExecuteInit initializes a file upload and returns an upload ID
func (uc *fileUploadUseCase) ExecuteInit(ctx context.Context, input FileUploadUseCaseExecuteInitInput) (*entity.File, []*entity.FileChunk, e.CustomError) {
	// check if the total size can fit in the available storage space
	availableSpace := uc.storageRepo.GetAvailableSpace(ctx, uc.baseStorageDir)
	if availableSpace < input.TotalSize {
		return nil, nil, e.NewFileStorageError(
			fmt.Errorf("not enough space"),
			fmt.Sprintf("File size of %s is %d bytes, but available space is %d bytes", input.FileName, input.TotalSize, availableSpace),
		)
	}

	// Directory path for the file chunks
	fileDirPath := filepath.Join(uc.baseStorageDir, input.FileName)

	// Check if the file already exists
	existingFile, err := uc.fileRepo.GetFileByName(ctx, input.FileName)
	if err != nil {
		return nil, nil, err
	}
	if existingFile == nil {
		// Create a new file object and insert it to "files" and corresponding child records"file_chunks" tables
		file := entity.NewFile(input.FileName, input.TotalSize, input.Checksum, input.TotalChunks, input.ChunkSize)

		fileRecord, err := uc.fileRepo.CreateFileWithChunks(ctx, file, uc.baseStorageDir)
		if err != nil {
			return nil, nil, err
		}

		// Create a directory for the file chunks
		if err := uc.storageRepo.CreateDirectory(ctx, fileDirPath); err != nil {
			return nil, nil, err
		}

		return fileRecord, nil, nil
	}

	// Assume two files are same if they have the same checksum and size
	isSameFile := existingFile.Checksum == input.Checksum && existingFile.Size == input.TotalSize
	if !isSameFile {
		return nil, nil, e.NewInvalidInputError(err, fmt.Sprintf("%s with different content (including orphaned data) already exists", input.FileName))
	}
	if existingFile.Status == entity.FileStatusInitialized {
		// Due to the atomicity of the transaction, if the status of "files" record is "INITIALIZED", that of "file_chunks" records is also
		// "INITIALIZED", so there's no need to check the inconsistency of file status and file chunk status. However, the file storage could
		// fail to be created even though the "files" record is created, so the following makes sure the directory is certainly there in the file storage
		if err := uc.storageRepo.CreateDirectory(ctx, fileDirPath); err != nil {
			return nil, nil, err
		}
		return existingFile, nil, nil
	}
	if !input.IsReUpload {
		return nil, nil, e.NewInvalidInputError(err, fmt.Sprintf("%s with same content(including orphaned data) already exists", input.FileName))
	}
	if existingFile.Status == entity.FileStatusUploaded {
		// only accepts re-uploading for files which aren't completed yet
		return nil, nil, e.NewInvalidInputError(err, fmt.Sprintf("existing %s is in %s status and cannot be re-uploaded", input.FileName, existingFile.Status))
	}

	// remove corresponding file chunks whose status is not "UPLOADED"
	invalidChunks, err := uc.fileRepo.GetChunksByStatus(ctx, existingFile.ID, []entity.FileStatus{
		entity.FileStatusInitialized,
		entity.FileStatusInProgress,
		entity.FileStatusFailed,
	})
	if err != nil {
		return nil, nil, err
	}
	// TODO: invalidChunks could be empty due to the failure in completed_chunks counter increment or status update
	for _, chunk := range invalidChunks {
		if err := uc.storageRepo.DeleteFile(ctx, chunk.FilePath); err != nil {
			return nil, nil, err
		}
	}

	// set the status of invalid chunks to "INITIALIZED"
	chunkIDsToUpdate := make([]uint64, 0, len(invalidChunks))
	for _, chunk := range invalidChunks {
		if chunk.Status != entity.FileStatusInitialized {
			chunkIDsToUpdate = append(chunkIDsToUpdate, chunk.ID)
		}
	}
	if len(chunkIDsToUpdate) > 0 {
		if err := uc.fileRepo.UpdateChunksStatus(ctx, chunkIDsToUpdate, entity.FileStatusInitialized); err != nil {
			return nil, nil, err
		}
	}

	// Preserving the space for the file to be uploaded
	uc.storageRepo.UpdateAvailableSpace(int64(input.TotalSize))

	return existingFile, invalidChunks, nil
}

// Execute uploads a chunk of a file
func (uc *fileUploadUseCase) Execute(ctx context.Context, input FileUploadUseCaseExecuteInput) e.CustomError {
	// validate the upload ID, chunk ID, and file and chunk status
	file, chunk, err := uc.fileRepo.GetFileAndChunk(ctx, input.FileID, input.ChunkNumber)
	if err != nil {
		return err
	}
	if file == nil || chunk == nil {
		return e.NewInvalidInputError(err, fmt.Sprintf("data not found for (file ID, chunk ID) = (%d, %d)", input.FileID, input.ChunkNumber))
	}
	if (file.Status != entity.FileStatusInitialized && file.Status != entity.FileStatusInProgress) || chunk.Status != entity.FileStatusInitialized {
		return e.NewInvalidInputError(fmt.Errorf("file upload needs to be initialized"), "")
	}

	// Update file and chunk status to processing
	if file.Status == entity.FileStatusInitialized {
		err = uc.fileRepo.UpdateFileAndChunkStatus(ctx, file.ID, chunk.ID, entity.FileStatusInProgress)
		if err != nil {
			return err
		}
	} else {
		// File status can be "IN_PROGRESS" or "FAILED" (due to other concurrent chunk uploads) at the time of this chunk upload. This chunk upload should
		// continue even though other concurrent chunks are failed to uploaded.
		err = uc.fileRepo.UpdateChunksStatus(ctx, []uint64{chunk.ID}, entity.FileStatusInProgress)
		if err != nil {
			return err
		}
	}

	// Write the chunk to storage
	if err := uc.storageRepo.WriteChunk(ctx, input.Reader, chunk.FilePath); err != nil {
		return err
	}

	// Update chunk status to completed
	if err = uc.fileRepo.UpdateChunksStatus(ctx, []uint64{chunk.ID}, entity.FileStatusUploaded); err != nil {
		return err
	}

	// Increment uploaded chunks counter
	uploadedChunks, totalChunks, err := uc.fileRepo.IncrementUploadedChunks(ctx, file.ID)
	if err != nil {
		return err
	}

	// Check if all chunks are uploaded
	if uploadedChunks == totalChunks {
		// Update file status to completed
		if err := uc.fileRepo.UpdateFileStatus(ctx, file.ID, entity.FileStatusUploaded); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteFailRecovery handles the failure of a chunk upload
func (uc *fileUploadUseCase) ExecuteFailRecovery(ctx context.Context, fileID uint64, chunkID uint64) e.CustomError {
	if err := uc.fileRepo.UpdateFileAndChunkStatus(ctx, fileID, chunkID, entity.FileStatusFailed); err != nil {
		return err
	}
	return nil
}
