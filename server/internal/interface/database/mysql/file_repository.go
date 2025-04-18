package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/core/entity"
	"github.com/tomoya.tokunaga/server/internal/core/repository"
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
func (r *fileRepository) CreateFile(ctx context.Context, file *entity.File) (uint64, error) {
	var model FileModel
	model.FromEntity(file)

	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Create(&model).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create file: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return model.ID, nil
}

// GetFileByID retrieves a file by its ID
func (r *fileRepository) GetFileByID(ctx context.Context, id uint64) (*entity.File, error) {
	var model FileModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file by ID: %w", err)
	}

	return model.ToEntity(), nil
}

// GetFileByName retrieves a file by its name
func (r *fileRepository) GetFileByName(ctx context.Context, name string) (*entity.File, error) {
	var model FileModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file by name: %w", err)
	}

	return model.ToEntity(), nil
}

// UpdateFileStatus updates the status of a file
func (r *fileRepository) UpdateFileStatus(ctx context.Context, id uint64, status entity.FileStatus) error {
	if err := r.db.WithContext(ctx).Model(&FileModel{}).Where("id = ?", id).Update("status", string(status)).Error; err != nil {
		return fmt.Errorf("failed to update file status: %w", err)
	}

	return nil
}

// IncrementCompletedChunks increments the completed chunks counter of a file
func (r *fileRepository) IncrementCompletedChunks(ctx context.Context, id uint64) (uint, uint, error) {
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Model(&FileModel{}).Where("id = ?", id).Update("completed_chunks", gorm.Expr("completed_chunks + ?", 1)).Error; err != nil {
		tx.Rollback()
		return 0, 0, fmt.Errorf("failed to increment completed chunks: %w", err)
	}

	var model FileModel
	if err := tx.Select("completed_chunks, total_chunks").First(&model, id).Error; err != nil {
		tx.Rollback()
		return 0, 0, fmt.Errorf("failed to get updated file: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return model.CompletedChunks, model.TotalChunks, nil
}

// DeleteFile deletes a file from the repository
func (r *fileRepository) DeleteFile(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&FileModel{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFiles lists all completed files
func (r *fileRepository) ListFiles(ctx context.Context) ([]string, error) {
	var fileNames []string
	if err := r.db.WithContext(ctx).Model(&FileModel{}).Where("status = ?", string(entity.FileStatusCompleted)).Pluck("name", &fileNames).Error; err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return fileNames, nil
}
