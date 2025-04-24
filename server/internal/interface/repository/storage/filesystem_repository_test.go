package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/exp/slog"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
)

func TestStorageRepository_WriteChunk(t *testing.T) {
	ctx := context.Background()
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()

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
				filePath := filepath.Join(tempDir, "subdir", "chunk_0")
				return filePath, func() {}
			},
			wantErr: false,
			checkResult: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				assert.Equal(t, "hello world", string(data))

				dirPath := filepath.Dir(filePath)
				_, err = os.Stat(dirPath)
				assert.NoError(t, err, "Directory should exist")
			},
		},
		{
			name:    "Success - Overwrite existing chunk",
			content: "new content",
			setup: func(t *testing.T) (string, func()) {
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
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()
	tests := []struct {
		name      string
		setup     func(t *testing.T) (filePath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, filePath string)
	}{
		{
			name: "Success - Delete existing chunk",
			setup: func(t *testing.T) (string, func()) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setup(t)
			defer cleanup()

			gotErr := repo.DeleteFile(ctx, filePath)

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
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()
	tests := []struct {
		name      string
		setup     func(t *testing.T) (dirPath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, dirPath string)
	}{
		{
			name: "Success - Create new directory",
			setup: func(t *testing.T) (string, func()) {
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
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()
	tests := []struct {
		name      string
		setup     func(t *testing.T) (dirPath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, dirPath string)
	}{
		{
			name: "Success - Delete empty directory",
			setup: func(t *testing.T) (string, func()) {
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
			defer cleanup()

			gotErr := repo.DeleteDirectory(ctx, dirPath)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, dirPath)
			}
			_ = cleanup
		})
	}
}

func TestStorageRepository_FileExists(t *testing.T) {
	ctx := context.Background()
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		setup      func(t *testing.T) (filePath string, cleanup func())
		wantExists bool
		wantErr    bool
	}{
		{
			name: "Success - File exists",
			setup: func(t *testing.T) (string, func()) {
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
				filePath := filepath.Join(tempDir, "non_existent_file.txt")
				return filePath, func() {}
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "Success - Path is a directory", // os.Stat returns info for directories too
			setup: func(t *testing.T) (string, func()) {
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
	config := entity.NewConfig()
	repo := NewStorageRepository(config, slog.Default())
	tempDir := t.TempDir()
	tests := []struct {
		name      string
		setup     func(t *testing.T) (filePath string, cleanup func())
		wantErr   bool
		checkFunc func(t *testing.T, filePath string)
	}{
		{
			name: "Success - Delete existing file",
			setup: func(t *testing.T) (string, func()) {
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

func TestStorageRepository_GetAvailableSpace(t *testing.T) {
	ctx := context.Background()
	config := entity.NewConfig()
	tempDir := t.TempDir()

	// Ensure the directory exists before creating the repository
	_, err := os.Stat(tempDir)
	assert.NoError(t, err, "Temp directory should exist")

	repo := NewStorageRepository(config, slog.Default())

	// Test getting the available space
	available := repo.GetAvailableSpace(ctx, tempDir)
	assert.NotZero(t, available, "Available space should not be zero")
}

func TestStorageRepository_UpdateAvailableSpace(t *testing.T) {
	ctx := context.Background()
	config := entity.NewConfig()
	tempDir := t.TempDir()
	repo := NewStorageRepository(config, slog.Default())

	// Get initial available space
	initialSpace := repo.GetAvailableSpace(ctx, tempDir)

	// Test reducing space
	sizeReduction := int64(1024 * 1024) // 1MB
	repo.UpdateAvailableSpace(-sizeReduction)

	// Test space was reduced
	reducedSpace := repo.GetAvailableSpace(ctx, tempDir)
	assert.Equal(t, initialSpace-uint64(sizeReduction), reducedSpace, "Space should be reduced by 1MB")

	// Test increasing space
	repo.UpdateAvailableSpace(sizeReduction)

	// Verify space was restored
	restoredSpace := repo.GetAvailableSpace(ctx, tempDir)
	assert.Equal(t, initialSpace, restoredSpace, "Space should be restored to initial value")

	// Test underflow protection
	hugeReduction := int64(initialSpace * 2) // More than available
	repo.UpdateAvailableSpace(-hugeReduction)

	// Test space is zero but doesn't underflow
	finalSpace := repo.GetAvailableSpace(ctx, tempDir)
	assert.Zero(t, finalSpace, "Space should be zero after large reduction")
}

func TestStorageRepository_InitWithNonExistentDirectory(t *testing.T) {
	tempBase := t.TempDir()
	nonExistentDir := filepath.Join(tempBase, "storage", "uploads")

	// Test directory doesn't exist yet
	_, err := os.Stat(nonExistentDir)
	assert.True(t, os.IsNotExist(err), "Directory should not exist before test")

	// Create a custom config that uses our test directory
	config := entity.NewConfig()
	config.BaseStorageDir = tempBase

	// Create repository - this should create the base directory
	repo := NewStorageRepository(config, slog.Default())

	// Now create the uploads subdirectory using the repo
	err = repo.CreateDirectory(context.Background(), nonExistentDir)
	assert.NoError(t, err, "Should be able to create the directory")

	// Test directory was created
	dirInfo, err := os.Stat(nonExistentDir)
	assert.NoError(t, err, "Directory should exist after repository creation")
	assert.True(t, dirInfo.IsDir(), "Path should be a directory")

	// Test we can get available space from the new directory
	available := repo.GetAvailableSpace(context.Background(), nonExistentDir)
	assert.NotZero(t, available, "Available space should not be zero")
}
