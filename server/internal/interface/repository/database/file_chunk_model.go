package database

import (
	"time"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
)

// FileChunkModel represents the file_chunks table in the database
type FileChunkModel struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	ParentID    uint64 `gorm:"index;not null"`
	Status      string `gorm:"type:enum('INITIALIZED','IN_PROGRESS','FAILED','UPLOADED');default:'INITIALIZED';not null"`
	ChunkNumber uint64 `gorm:"not null;default:0"`
	Size        uint64 `gorm:"not null;default:0"`
	FilePath    string `gorm:"size:1024;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName returns the table name for the file chunk model
func (FileChunkModel) TableName() string {
	return "file_chunks"
}

// ToEntity converts a FileChunkModel to a FileChunk entity
func (m *FileChunkModel) ToEntity() *entity.FileChunk {
	return &entity.FileChunk{
		ID:          m.ID,
		ParentID:    m.ParentID,
		Status:      entity.FileStatus(m.Status),
		ChunkNumber: m.ChunkNumber,
		Size:        m.Size,
		FilePath:    m.FilePath,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromEntity converts a FileChunk entity to a FileChunkModel
func (m *FileChunkModel) FromEntity(e *entity.FileChunk) {
	m.ID = e.ID
	m.ParentID = e.ParentID
	m.Status = string(e.Status)
	m.ChunkNumber = e.ChunkNumber
	m.Size = e.Size
	m.FilePath = e.FilePath
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}
