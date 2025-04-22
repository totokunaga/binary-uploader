package database

import (
	"time"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
)

// FileModel represents the files table in the database
type FileModel struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement"`
	Name           string `gorm:"uniqueIndex;size:255;not null"`
	Size           uint64 `gorm:"not null;default:0"`
	Checksum       string `gorm:"size:512;not null"`
	ChunkSize      uint64 `gorm:"not null;default:0"`
	Status         string `gorm:"type:enum('INITIALIZED','IN_PROGRESS','FAILED','UPLOADED');default:'INITIALIZED';index;not null"`
	TotalChunks    uint   `gorm:"not null;default:0"`
	UploadedChunks uint   `gorm:"not null;default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	FileChunks     []FileChunkModel `gorm:"foreignKey:ParentID"`
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
		ID:             m.ID,
		Name:           m.Name,
		Size:           m.Size,
		Checksum:       m.Checksum,
		ChunkSize:      m.ChunkSize,
		Status:         entity.FileStatus(m.Status),
		TotalChunks:    m.TotalChunks,
		UploadedChunks: m.UploadedChunks,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		FileChunks:     fileChunks,
	}
}

// FromEntity converts a File entity to a FileModel
func (m *FileModel) FromEntity(e *entity.File) {
	m.ID = e.ID
	m.Name = e.Name
	m.Size = e.Size
	m.Checksum = e.Checksum
	m.ChunkSize = e.ChunkSize
	m.Status = string(e.Status)
	m.TotalChunks = e.TotalChunks
	m.UploadedChunks = e.UploadedChunks
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}
