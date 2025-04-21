package database

import (
	"time"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
)

// FileChunkModel represents the file_chunks table in the database
type FileChunkModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ParentID  uint64 `gorm:"index;not null"`
	Status    string `gorm:"type:enum('UPLOAD_INITIALIZED','UPLOAD_IN_PROGRESS','UPLOAD_FAILED','UPLOADED','DELETE_INITIALIZED','DELETE_IN_PROGRESS','DELETE_FAILED','DELETED');default:'UPLOAD_INITIALIZED';not null"`
	FilePath  string `gorm:"size:1024;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the file chunk model
func (FileChunkModel) TableName() string {
	return "file_chunks"
}

// ToEntity converts a FileChunkModel to a FileChunk entity
func (m *FileChunkModel) ToEntity() *entity.FileChunk {
	return &entity.FileChunk{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Status:    entity.FileChunkStatus(m.Status),
		FilePath:  m.FilePath,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromEntity converts a FileChunk entity to a FileChunkModel
func (m *FileChunkModel) FromEntity(e *entity.FileChunk) {
	m.ID = e.ID
	m.ParentID = e.ParentID
	m.Status = string(e.Status)
	m.FilePath = e.FilePath
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}
