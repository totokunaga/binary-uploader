package database

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// fileRepository implements the repository.FileRepository interface
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new MySQL file repository
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

// GetFileByName retrieves a file by its name
func (r *fileRepository) GetFileByName(ctx context.Context, name string) (*entity.File, e.CustomError) {
	var model FileModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e.NewDatabaseError(err, "GetFileByName: failed to get file by name")
	}

	return model.ToEntity(), nil
}

// GetChunksByFileID retrieves all file chunks associated with a given file ID.
func (r *fileRepository) GetChunksByFileID(ctx context.Context, fileID uint64) ([]*entity.FileChunk, e.CustomError) {
	var chunkModels []FileChunkModel

	err := r.db.WithContext(ctx).
		Where("parent_id = ?", fileID).
		Find(&chunkModels).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.FileChunk{}, nil
		}
		return nil, e.NewDatabaseError(err, "GetChunksByFileID: failed to get file chunks by file ID")
	}

	chunks := make([]*entity.FileChunk, len(chunkModels))
	for i, model := range chunkModels {
		chunks[i] = model.ToEntity()
		if chunks[i] == nil {
			return nil, e.NewDatabaseError(fmt.Errorf("failed to convert chunk model ID %d to entity", model.ID), "GetChunksByFileID: chunk conversion failed")
		}
	}

	return chunks, nil
}

// GetChunksByStatus retrieves file chunks associated with a given file ID that match any of the specified statuses.
func (r *fileRepository) GetChunksByStatus(ctx context.Context, fileID uint64, statuses []entity.FileStatus) ([]*entity.FileChunk, e.CustomError) {
	var chunkModels []FileChunkModel
	statusStrings := make([]string, len(statuses))
	for i, s := range statuses {
		statusStrings[i] = string(s)
	}

	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND status IN ?", fileID, statusStrings).
		Find(&chunkModels).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.FileChunk{}, nil
		}
		return nil, e.NewDatabaseError(err, "GetChunksByStatus: failed to get file chunks by status")
	}

	// Convert models to entities
	chunks := make([]*entity.FileChunk, len(chunkModels))
	for i, model := range chunkModels {
		chunks[i] = model.ToEntity()
		if chunks[i] == nil {
			return nil, e.NewDatabaseError(fmt.Errorf("failed to convert chunk model ID %d to entity", model.ID), "GetChunksByStatus: chunk conversion failed")
		}
	}

	return chunks, nil
}

// GetFileAndChunk retrieves a file and a specific chunk by file ID and chunk number.
func (r *fileRepository) GetFileAndChunk(ctx context.Context, fileID uint64, chunkNumber uint64) (*entity.File, *entity.FileChunk, e.CustomError) {
	var fileModel FileModel
	var chunkModel FileChunkModel

	db := r.db.WithContext(ctx)

	// Get the file
	if err := db.First(&fileModel, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, e.NewNotFoundError(err, fmt.Sprintf("file with ID %d not found", fileID))
		}
		return nil, nil, e.NewDatabaseError(err, "GetFileAndChunk: failed to get file by ID")
	}

	// Get the specific chunk
	if err := db.Where("parent_id = ? AND chunk_number = ?", fileID, chunkNumber).First(&chunkModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, e.NewNotFoundError(err, fmt.Sprintf("chunk number %d for file ID %d not found", chunkNumber, fileID))
		}
		return nil, nil, e.NewDatabaseError(err, "GetFileAndChunk: failed to get file chunk")
	}

	fileEntity := fileModel.ToEntity()
	if fileEntity == nil {
		return nil, nil, e.NewDatabaseError(errors.New("failed to convert file model to entity"), "GetFileAndChunk: file conversion failed")
	}

	chunkEntity := chunkModel.ToEntity()
	if chunkEntity == nil {
		return nil, nil, e.NewDatabaseError(errors.New("failed to convert chunk model to entity"), "GetFileAndChunk: chunk conversion failed")
	}

	return fileEntity, chunkEntity, nil
}

// GetFileNames lists all completed files
func (r *fileRepository) GetFileNames(ctx context.Context) ([]string, e.CustomError) {
	var fileNames []string

	err := r.db.WithContext(ctx).Model(&FileModel{}).
		Where("status = ?", string(entity.FileStatusUploaded)).
		Pluck("name", &fileNames).Error

	if err != nil {
		return nil, e.NewDatabaseError(err, "ListFiles: failed to list files")
	}

	return fileNames, nil
}

// CreateFileWithChunks creates a file and its corresponding chunk records within a transaction.
func (r *fileRepository) CreateFileWithChunks(ctx context.Context, file *entity.File, baseDir string) (*entity.File, e.CustomError) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, e.NewDatabaseError(tx.Error, "CreateFileWithChunks: failed to begin transaction")
	}

	// Create FileModel
	var fileModel FileModel
	fileModel.FromEntity(file)
	if err := tx.Create(&fileModel).Error; err != nil {
		tx.Rollback()
		return nil, e.NewDatabaseError(err, "CreateFileWithChunks: failed to create file record")
	}

	// Create FileChunkModels
	fileChunks := make([]FileChunkModel, file.TotalChunks)
	for i := uint(0); i < file.TotalChunks; i++ {
		fileChunks[i] = FileChunkModel{
			ParentID:    fileModel.ID,
			ChunkNumber: uint64(i),
			Status:      string(entity.FileStatusInitialized),
			FilePath:    filepath.Join(baseDir, file.Name, fmt.Sprintf("%d", i)),
		}
	}

	// Use CreateInBatches to handle potentially large numbers of chunks
	if err := tx.CreateInBatches(&fileChunks, 1000).Error; err != nil {
		tx.Rollback()
		return nil, e.NewDatabaseError(err, "CreateFileWithChunks: failed to create file chunks")
	}

	if err := tx.Commit().Error; err != nil {
		// Rollback might have already happened implicitly, but doesn't hurt.
		tx.Rollback()
		return nil, e.NewDatabaseError(err, "CreateFileWithChunks: failed to commit transaction")
	}

	createdFile := fileModel.ToEntity()
	if createdFile == nil {
		// Handle the case where conversion fails, although it shouldn't if creation succeeded
		return nil, e.NewDatabaseError(errors.New("failed to convert created file model to entity"), "CreateFileWithChunks: database conversion error")
	}

	return createdFile, nil
}

// DeleteFileByID deletes a file by its ID. This also triggers the cascade delete of all its chunks.
func (r *fileRepository) DeleteFileByID(ctx context.Context, fileID uint64) e.CustomError {
	if err := r.db.WithContext(ctx).Delete(&FileModel{}, fileID).Error; err != nil {
		return e.NewDatabaseError(err, "DeleteFileByID: failed to delete file")
	}
	return nil
}

// UpdateFileStatus updates the status of a file
func (r *fileRepository) UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) e.CustomError {
	if err := r.db.WithContext(ctx).Model(&FileModel{}).Where("id = ?", id).Update("status", string(status)).Error; err != nil {
		return e.NewDatabaseError(err, "UpdateFileStatus: failed to update file status")
	}

	return nil
}

// UpdateChunksStatus updates the status of multiple file chunks identified by their IDs within a transaction.
func (r *fileRepository) UpdateChunksStatus(ctx context.Context, chunkIDs []uint64, status entity.FileStatus) e.CustomError {
	if len(chunkIDs) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return e.NewDatabaseError(tx.Error, "UpdateChunksStatus: failed to begin transaction")
	}

	// Update chunk statuses
	if err := tx.Model(&FileChunkModel{}).Where("id IN ?", chunkIDs).Update("status", string(status)).Error; err != nil {
		tx.Rollback()
		return e.NewDatabaseError(err, "UpdateChunksStatus: failed to update file chunk statuses")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return e.NewDatabaseError(err, "UpdateChunksStatus: failed to commit transaction")
	}

	return nil
}

// UpdateFileAndChunkStatus updates the status of a specific file and a specific chunk within a transaction.
func (r *fileRepository) UpdateFileAndChunkStatus(ctx context.Context, fileID uint64, chunkIDs []uint64, status entity.FileStatus) e.CustomError {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return e.NewDatabaseError(tx.Error, "UpdateFileAndChunkStatus: failed to begin transaction")
	}

	statusStr := string(status)

	// Update file status
	if err := tx.Model(&FileModel{}).Where("id = ?", fileID).Update("status", statusStr).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return e.NewNotFoundError(err, fmt.Sprintf("UpdateFileAndChunkStatus: file with ID %d not found for status update", fileID))
		}
		return e.NewDatabaseError(err, "UpdateFileAndChunkStatus: failed to update file status")
	}

	// Update chunk status
	if err := tx.Model(&FileChunkModel{}).Where("id IN ?", chunkIDs).Update("status", statusStr).Error; err != nil {
		tx.Rollback()
		return e.NewDatabaseError(err, fmt.Sprintf("UpdateFileAndChunkStatus: failed to update file chunk statuses for IDs %v", chunkIDs))
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return e.NewDatabaseError(err, "UpdateFileAndChunkStatus: failed to commit transaction")
	}

	return nil
}

// IncrementUploadedChunks increments the uploaded chunks counter of a file
func (r *fileRepository) IncrementUploadedChunks(ctx context.Context, id uint64) (uint, uint, e.CustomError) {
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Model(&FileModel{}).Where("id = ?", id).Update("uploaded_chunks", gorm.Expr("uploaded_chunks + ?", 1)).Error; err != nil {
		tx.Rollback()
		return 0, 0, e.NewDatabaseError(err, "IncrementUploadedChunks: failed to increment uploaded chunks")
	}

	var model FileModel
	if err := tx.Select("uploaded_chunks, total_chunks").First(&model, id).Error; err != nil {
		tx.Rollback()
		return 0, 0, e.NewDatabaseError(err, "IncrementUploadedChunks: failed to get updated file")
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, e.NewDatabaseError(err, "IncrementUploadedChunks: failed to commit transaction")
	}

	return model.UploadedChunks, model.TotalChunks, nil
}

// CountChunksByStatus counts the number of chunks for a file with a specific status
// and returns the total number of chunks for that file.
func (r *fileRepository) CountChunksByStatus(ctx context.Context, fileID uint64, status entity.FileStatus) (int64, int64, e.CustomError) {
	var count int64
	var fileModel FileModel
	statusStr := string(status)
	db := r.db.WithContext(ctx)

	if err := db.Select("total_chunks").First(&fileModel, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, e.NewNotFoundError(err, fmt.Sprintf("CountChunksByStatus: file with ID %d not found", fileID))
		}
		return 0, 0, e.NewDatabaseError(err, "CountChunksByStatus: failed to get file")
	}

	err := db.
		Model(&FileChunkModel{}).
		Where("parent_id = ? AND status = ?", fileID, statusStr).
		Count(&count).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, int64(fileModel.TotalChunks), e.NewDatabaseError(err, fmt.Sprintf("CountChunksByStatus: failed to count chunks for file ID %d with status %s", fileID, statusStr))
		}
	}

	return count, int64(fileModel.TotalChunks), nil
}
