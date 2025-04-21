package database

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
)

// fileRepository implements the repository.FileRepository interface
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new MySQL file repository
func NewFileRepository(db *gorm.DB) repository.FileRepository {
	return &fileRepository{db: db}
}

// CreateFile creates a new file entry in the repository
func (r *fileRepository) CreateFile(ctx context.Context, file *entity.File) (uint64, e.CustomError) {
	var model FileModel
	model.FromEntity(file)

	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Create(&model).Error; err != nil {
		tx.Rollback()
		return 0, e.NewDatabaseError(err, "failed to create file record")
	}
	if err := tx.Commit().Error; err != nil {
		return 0, e.NewDatabaseError(err, "failed to commit transaction")
	}

	return model.ID, nil
}

// GetFileByID retrieves a file by its ID
func (r *fileRepository) GetFileByID(ctx context.Context, id uint64) (*entity.File, e.CustomError) {
	var model FileModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e.NewDatabaseError(err, "failed to get file by ID")
	}

	return model.ToEntity(), nil
}

// GetFileByName retrieves a file by its name
func (r *fileRepository) GetFileByName(ctx context.Context, name string) (*entity.File, e.CustomError) {
	var model FileModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e.NewDatabaseError(err, "failed to get file by name")
	}

	return model.ToEntity(), nil
}

// UpdateFileStatus updates the status of a file
func (r *fileRepository) UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) e.CustomError {
	if err := r.db.WithContext(ctx).Model(&FileModel{}).Where("id = ?", id).Update("status", string(status)).Error; err != nil {
		return e.NewDatabaseError(err, "failed to update file status")
	}

	return nil
}

// IncrementCompletedChunks increments the completed chunks counter of a file
func (r *fileRepository) IncrementCompletedChunks(ctx context.Context, id uint64) (uint, uint, e.CustomError) {
	// TODO: should use mutex?
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Model(&FileModel{}).Where("id = ?", id).Update("completed_chunks", gorm.Expr("completed_chunks + ?", 1)).Error; err != nil {
		tx.Rollback()
		return 0, 0, e.NewDatabaseError(err, "failed to increment completed chunks")
	}

	var model FileModel
	if err := tx.Select("completed_chunks, total_chunks").First(&model, id).Error; err != nil {
		tx.Rollback()
		return 0, 0, e.NewDatabaseError(err, "failed to get updated file")
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, e.NewDatabaseError(err, "failed to commit transaction")
	}

	return model.CompletedChunks, model.TotalChunks, nil
}

// DeleteFile deletes a file from the repository
func (r *fileRepository) DeleteFile(ctx context.Context, id uint64) e.CustomError {
	if err := r.db.WithContext(ctx).Delete(&FileModel{}, id).Error; err != nil {
		return e.NewDatabaseError(err, "failed to delete file")
	}

	return nil
}

// ListFiles lists all completed files
func (r *fileRepository) ListFiles(ctx context.Context) ([]string, e.CustomError) {
	var fileNames []string

	err := r.db.WithContext(ctx).Model(&FileModel{}).
		Where("status = ?", string(entity.FileStatusUploaded)).
		Pluck("name", &fileNames).Error

	if err != nil {
		return nil, e.NewDatabaseError(err, "failed to list files")
	}

	return fileNames, nil
}
