package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/exp/slog"

	"github.com/stretchr/testify/assert"
)

func TestStorageRepository_WriteChunk(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name        string
		content     string
		setup       func(t *testing.T) (filePath string, cleanup func())
		wantErr     bool
		checkResult func(t *testing.T, filePath string)
	}{
		{
			name:    "Success - Write new chunk",
			content: "hello world",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "subdir", "chunk_0")
				return filePath, func() { /* TempDir handles cleanup */ }
			},
			wantErr: false,
			checkResult: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				assert.Equal(t, "hello world", string(data))
				// Check if directory was created
				dirPath := filepath.Dir(filePath)
				_, err = os.Stat(dirPath)
				assert.NoError(t, err, "Directory should exist")
			},
		},
		{
			name:    "Success - Overwrite existing chunk",
			content: "new content",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "chunk_overwrite")
				err := os.WriteFile(filePath, []byte("old content"), 0644)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantErr: false,
			checkResult: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				assert.Equal(t, "new content", string(data))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setup(t)
			defer cleanup()

			reader := strings.NewReader(tt.content)
			gotErr := repo.WriteChunk(ctx, reader, filePath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, filePath)
			}
		})
	}
}

func TestStorageRepository_DeleteChunk(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name      string
		setup     func(t *testing.T) (filePath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, filePath string)
	}{
		{
			name: "Success - Delete existing chunk",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "chunk_to_delete")
				err := os.WriteFile(filePath, []byte("data"), 0644)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, filePath string) {
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err), "File should not exist after deletion")
			},
		},
		{
			name: "Success - Delete non-existent chunk",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "non_existent_chunk")
				return filePath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, filePath string) {
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
		// Note: Testing actual permission errors reliably might require more complex setup
		// or specific OS environments.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setup(t)
			defer cleanup()

			gotErr := repo.DeleteChunk(ctx, filePath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, filePath)
			}
		})
	}
}

func TestStorageRepository_CreateDirectory(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name      string
		setup     func(t *testing.T) (dirPath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, dirPath string)
	}{
		{
			name: "Success - Create new directory",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "new_dir", "subdir")
				return dirPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir(), "Path should be a directory")
			},
		},
		{
			name: "Success - Create existing directory",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "existing_dir")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				return dirPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Failure - Path is a file",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "existing_file")
				_, err := os.Create(filePath)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantErr: true,
			checkFunc: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err) // File should still exist
				assert.False(t, info.IsDir(), "Path should remain a file")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath, cleanup := tt.setup(t)
			defer cleanup()

			gotErr := repo.CreateDirectory(ctx, dirPath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, dirPath)
			}
		})
	}
}

func TestStorageRepository_DeleteDirectory(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name      string
		setup     func(t *testing.T) (dirPath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, dirPath string)
	}{
		{
			name: "Success - Delete empty directory",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "empty_dir_to_delete")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				return dirPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				_, err := os.Stat(dirPath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err), "Directory should not exist")
			},
		},
		{
			name: "Success - Delete directory with contents",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "dir_with_content")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				filePath := filepath.Join(dirPath, "somefile.txt")
				err = os.WriteFile(filePath, []byte("content"), 0644)
				assert.NoError(t, err)
				return dirPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				_, err := os.Stat(dirPath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err), "Directory should not exist")
			},
		},
		{
			name: "Success - Delete non-existent directory",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "non_existent_dir")
				return dirPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				_, err := os.Stat(dirPath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "Failure - Path is a file",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "a_file_not_a_dir")
				_, err := os.Create(filePath)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, dirPath string) {
				_, err := os.Stat(dirPath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath, cleanup := tt.setup(t)
			// No defer cleanup() here because the function under test might delete the path

			gotErr := repo.DeleteDirectory(ctx, dirPath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, dirPath)
			}
			// Manually call cleanup if the test didn't expect deletion or failed before deletion
			// This is tricky. t.TempDir() handles cleanup usually.
			// For manual paths, cleanup might be needed conditionally.
			// Sticking with TempDir avoids this complexity.
			_ = cleanup // Use cleanup if manual paths were used and needed conditional cleanup
		})
	}
}

func TestStorageRepository_FileExists(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name       string
		setup      func(t *testing.T) (filePath string, cleanup func())
		wantExists bool
		wantErr    bool
	}{
		{
			name: "Success - File exists",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "existing_file.txt")
				_, err := os.Create(filePath)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name: "Success - File does not exist",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "non_existent_file.txt")
				return filePath, func() {}
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "Success - Path is a directory", // os.Stat returns info for directories too
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "a_directory")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				return dirPath, func() {}
			},
			wantExists: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setup(t)
			defer cleanup()

			gotExists, gotErr := repo.FileExists(ctx, filePath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.Equal(t, tt.wantExists, gotExists)

		})
	}
}

func TestStorageRepository_DeleteFile(t *testing.T) {
	ctx := context.Background()
	repo := NewStorageRepository(slog.Default())

	tests := []struct {
		name      string
		setup     func(t *testing.T) (filePath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, filePath string)
	}{
		{
			name: "Success - Delete existing file",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "file_to_delete")
				err := os.WriteFile(filePath, []byte("data"), 0644)
				assert.NoError(t, err)
				return filePath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, filePath string) {
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err), "File should not exist after deletion")
			},
		},
		{
			name: "Success - Delete non-existent file",
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "non_existent_file")
				return filePath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, filePath string) {
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "Failure - Path is a directory", // os.Remove fails on non-empty directories
			setup: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				dirPath := filepath.Join(tempDir, "a_dir_not_a_file")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				// Add content to make it non-empty
				err = os.WriteFile(filepath.Join(dirPath, "dummy.txt"), []byte("data"), 0644)
				assert.NoError(t, err)
				return dirPath, func() {}
			},
			wantErr: true, // os.Remove should error here
			checkFunc: func(t *testing.T, filePath string) {
				// Directory should still exist
				info, err := os.Stat(filePath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		// Note: Testing permission errors might require specific setup.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filePath, cleanup := tt.setup(t)
			// No defer cleanup here as the function might delete the path

			gotErr := repo.DeleteFile(ctx, filePath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, filePath)
			}
			_ = cleanup
		})
	}
}
