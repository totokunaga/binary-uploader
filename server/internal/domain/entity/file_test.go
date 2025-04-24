package entity

import (
	"testing"
)

func TestNewFile(t *testing.T) {
	name := "test-file.txt"
	size := uint64(1024)
	checksum := "abc123"
	totalChunks := uint(10)
	chunkSize := uint64(102)

	file := NewFile(name, size, checksum, totalChunks, chunkSize)

	if file.Name != name {
		t.Errorf("Expected Name to be %s, got %s", name, file.Name)
	}
	if file.Size != size {
		t.Errorf("Expected Size to be %d, got %d", size, file.Size)
	}
	if file.Checksum != checksum {
		t.Errorf("Expected Checksum to be %s, got %s", checksum, file.Checksum)
	}
	if file.TotalChunks != totalChunks {
		t.Errorf("Expected TotalChunks to be %d, got %d", totalChunks, file.TotalChunks)
	}
	if file.ChunkSize != chunkSize {
		t.Errorf("Expected ChunkSize to be %d, got %d", chunkSize, file.ChunkSize)
	}
	if file.Status != FileStatusInitialized {
		t.Errorf("Expected Status to be %s, got %s", FileStatusInitialized, file.Status)
	}
	if file.UploadedChunks != 0 {
		t.Errorf("Expected UploadedChunks to be 0, got %d", file.UploadedChunks)
	}
}

func TestFileStatusConstants(t *testing.T) {
	statuses := []struct {
		constant FileStatus
		expected string
	}{
		{FileStatusInitialized, "INITIALIZED"},
		{FileStatusInProgress, "IN_PROGRESS"},
		{FileStatusFailed, "FAILED"},
		{FileStatusUploaded, "UPLOADED"},
	}

	for _, s := range statuses {
		if string(s.constant) != s.expected {
			t.Errorf("Expected %s, got %s", s.expected, s.constant)
		}
	}
}
