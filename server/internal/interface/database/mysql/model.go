package mysql

import (
	"time"

	"github.com/tomoya.tokunaga/server/internal/core/entity"
)

// FileModel represents the files table in the database
type FileModel struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	Name            string `gorm:"uniqueIndex;size:255;not null"`
	Size            uint64 `gorm:"not null;default:0"`
	Status          string `gorm:"type:enum('INITIALIZED','PROCESSING','FAILED','COMPLETED');default:'INITIALIZED';index;not null"`
	TotalChunks     uint   `gorm:"not null;default:0"`
	CompletedChunks uint   `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FileChunks      []FileChunkModel `gorm:"foreignKey:ParentID"`
}

// TableName returns the table name for the file model
func (FileModel) TableName() string {
	return "files"
}

// ToEntity converts a FileModel to a File entity
func (m *FileModel) ToEntity() *entity.File {
	fileChunks := make([]entity.FileChunk, len(m.FileChunks))
	for i, chunk := range m.FileChunks {
		fileChunks[i] = *chunk.ToEntity()
	}

	return &entity.File{
		ID:              m.ID,
		Name:            m.Name,
		Size:            m.Size,
		Status:          entity.FileStatus(m.Status),
		TotalChunks:     m.TotalChunks,
		CompletedChunks: m.CompletedChunks,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		FileChunks:      fileChunks,
	}
}

// FromEntity converts a File entity to a FileModel
func (m *FileModel) FromEntity(e *entity.File) {
	m.ID = e.ID
	m.Name = e.Name
	m.Size = e.Size
	m.Status = string(e.Status)
	m.TotalChunks = e.TotalChunks
	m.CompletedChunks = e.CompletedChunks
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}

// FileChunkModel represents the file_chunks table in the database
type FileChunkModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ParentID  uint64 `gorm:"index;not null"`
	Status    string `gorm:"type:enum('INITIALIZED','PROCESSING','FAILED','COMPLETED');default:'INITIALIZED';not null"`
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
