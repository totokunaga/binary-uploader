package database

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
)

// fileChunkRepository implements the repository.FileChunkRepository interface
type fileChunkRepository struct {
	db *gorm.DB
}

// NewFileChunkRepository creates a new MySQL file chunk repository
func NewFileChunkRepository(db *gorm.DB) repository.FileChunkRepository {
	return &fileChunkRepository{db: db}
}

// CreateFileChunks creates new file chunk entries in the repository
func (r *fileChunkRepository) CreateFileChunks(ctx context.Context, fileID uint64, totalChunks uint, baseDir string, fileName string) e.CustomError {
	tx := r.db.WithContext(ctx).Begin()

	fileChunks := make([]FileChunkModel, totalChunks)
	for i := uint(0); i < totalChunks; i++ {
		fileChunks[i] = FileChunkModel{
			ParentID: fileID,
			Status:   string(entity.FileChunkStatusUploadInitialized),
			FilePath: filepath.Join(baseDir, fileName, fmt.Sprintf("%d", i)),
		}
	}

	if err := tx.Create(&fileChunks).Error; err != nil {
		tx.Rollback()
		return e.NewDatabaseError(err, "failed to create file chunks")
	}

	if err := tx.Commit().Error; err != nil {
		return e.NewDatabaseError(err, "failed to commit transaction")
	}

	return nil
}

// GetFileChunkByID retrieves a file chunk by its ID
func (r *fileChunkRepository) GetFileChunkByID(ctx context.Context, id uint64) (*entity.FileChunk, e.CustomError) {
	var model FileChunkModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e.NewDatabaseError(err, "failed to get file chunk by ID")
	}

	return model.ToEntity(), nil
}

// GetFileChunk retrieves a file chunk by parent ID and chunk ID
func (r *fileChunkRepository) GetFileChunk(ctx context.Context, parentID uint64, chunkID uint) (*entity.FileChunk, e.CustomError) {
	var model FileChunkModel
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Offset(int(chunkID)).Limit(1).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e.NewDatabaseError(err, "failed to get file chunk")
	}

	return model.ToEntity(), nil
}

// UpdateFileChunkStatus updates the status of a file chunk
func (r *fileChunkRepository) UpdateFileChunkStatus(ctx context.Context, id uint64, status entity.FileChunkStatus) e.CustomError {
	if err := r.db.WithContext(ctx).Model(&FileChunkModel{}).Where("id = ?", id).Update("status", string(status)).Error; err != nil {
		return e.NewDatabaseError(err, "failed to update file chunk status")
	}

	return nil
}

// UpdateFileChunkStatusByParentID updates the status of all file chunks for a given parent ID
func (r *fileChunkRepository) UpdateFileChunkStatusByParentID(ctx context.Context, parentID uint64, status entity.FileChunkStatus) e.CustomError {
	if err := r.db.WithContext(ctx).Model(&FileChunkModel{}).Where("parent_id = ?", parentID).Update("status", string(status)).Error; err != nil {
		return e.NewDatabaseError(err, "failed to update file chunk status")
	}

	return nil
}

// GetFileChunksByParentID retrieves all file chunks for a given parent ID
func (r *fileChunkRepository) GetFileChunksByParentID(ctx context.Context, parentID uint64) ([]entity.FileChunk, e.CustomError) {
	var models []FileChunkModel
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&models).Error; err != nil {
		return nil, e.NewDatabaseError(err, "failed to get file chunks by parent ID")
	}

	chunks := make([]entity.FileChunk, len(models))
	for i, model := range models {
		chunks[i] = *model.ToEntity()
	}

	return chunks, nil
}

// DeleteFileChunksByParentID deletes all file chunks for a given parent ID
func (r *fileChunkRepository) DeleteFileChunksByParentID(ctx context.Context, parentID uint64) e.CustomError {
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Delete(&FileChunkModel{}).Error; err != nil {
		return e.NewDatabaseError(err, "failed to delete file chunks by parent ID")
	}

	return nil
}
