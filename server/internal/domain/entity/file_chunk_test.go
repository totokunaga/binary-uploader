package entity

import (
	"testing"
)

func TestNewFileChunk(t *testing.T) {
	parentID := uint64(123)
	status := FileStatusInProgress
	chunkNumber := uint64(5)
	size := uint64(1024)
	filePath := "/path/to/chunk/file"

	chunk := NewFileChunk(parentID, status, chunkNumber, size, filePath)

	if chunk.ParentID != parentID {
		t.Errorf("Expected ParentID to be %d, got %d", parentID, chunk.ParentID)
	}
	if chunk.Status != status {
		t.Errorf("Expected Status to be %s, got %s", status, chunk.Status)
	}
	if chunk.ChunkNumber != chunkNumber {
		t.Errorf("Expected ChunkNumber to be %d, got %d", chunkNumber, chunk.ChunkNumber)
	}
	if chunk.Size != size {
		t.Errorf("Expected Size to be %d, got %d", size, chunk.Size)
	}
	if chunk.FilePath != filePath {
		t.Errorf("Expected FilePath to be %s, got %s", filePath, chunk.FilePath)
	}

	// Test ID and timestamps are not set by the constructor
	if chunk.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", chunk.ID)
	}
	if !chunk.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be zero time, got %v", chunk.CreatedAt)
	}
	if !chunk.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be zero time, got %v", chunk.UpdatedAt)
	}
}
