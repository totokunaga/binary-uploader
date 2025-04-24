package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
)

// storageRepository implements the repository.StorageRepository interface
type storageRepository struct {
	config *entity.Config
	logger *slog.Logger
	// stores the available space in bytes to tell if a new file upload can fit
	// in the storage (using syscall can be expensive, so in-memory is used)
	availableSpace uint64
	spaceMutex     sync.RWMutex
}

// NewStorageRepository creates a new filesystem storage repository
func NewStorageRepository(config *entity.Config, logger *slog.Logger) FileStorageRepository {
	repo := &storageRepository{
		logger: logger,
		config: config,
	}

	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(config.BaseStorageDir, 0755); err != nil {
		logger.Error("Failed to create base storage directory", "path", config.BaseStorageDir, "error", err)
		// when the directory is not created, the repository cannot work
		panic(err)
	}

	// Initialize available space
	var stat syscall.Statfs_t
	if err := syscall.Statfs(config.BaseStorageDir, &stat); err != nil {
		logger.Error("Failed to initialize available space", "path", config.BaseStorageDir, "error", err)
		// when the directory is not there or server cannot access it, the repository cannot work
		panic(err)
	}
	repo.availableSpace = stat.Bavail * uint64(stat.Bsize)

	return repo
}

// CreateDirectory creates a directory at the given path
func (r *storageRepository) CreateDirectory(ctx context.Context, dirPath string) e.CustomError {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return e.NewFileStorageError(err, "failed to create directory")
	}

	return nil
}

// WriteChunk writes the file chunk to the storage
func (r *storageRepository) WriteChunk(ctx context.Context, reader io.Reader, filePath string) e.CustomError {
	// Ensure the directory exists
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return e.NewFileStorageError(err, "failed to create directory")
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return e.NewFileStorageError(err, "failed to create file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			r.logger.Error("Failed to close file", "error", err)
		}
	}()

	// Read the chunk and write it to the file
	buffer := make([]byte, r.config.StreamBufferSize)
	for {
		select {
		case <-ctx.Done():
			return e.NewContextError(ctx.Err(), "context canceled")
		default:
			// Read a chunk from the reader
			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				return e.NewFileStorageError(err, "failed to read from reader")
			}
			if n == 0 {
				return nil
			}

			// Write the chunk to the file
			if _, err := file.Write(buffer[:n]); err != nil {
				return e.NewFileStorageError(err, "failed to write to file")
			}
		}
	}
}

// DeleteFile deletes a file from the storage
func (r *storageRepository) DeleteFile(ctx context.Context, filePath string) e.CustomError {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return e.NewFileStorageError(err, "failed to delete file")
	}

	return nil
}

// DeleteDirectory deletes a directory at the given path
func (r *storageRepository) DeleteDirectory(ctx context.Context, dirPath string) e.CustomError {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return e.NewFileStorageError(err, "failed to delete directory")
	}

	return nil
}

// // DeleteFile deletes a file at the given path
// func (r *storageRepository) DeleteFile(ctx context.Context, filePath string) e.CustomError {
// 	// Check if the file exists
// 	_, err := os.Stat(filePath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return nil
// 		}
// 		return e.NewFileStorageError(err, "failed to check file status before deletion")
// 	}

// 	// Delete the file
// 	if err := os.Remove(filePath); err != nil {
// 		return e.NewFileStorageError(err, "failed to delete file")
// 	}

// 	return nil
// }

// FileExists checks if a file exists at the given path
func (r *storageRepository) FileExists(ctx context.Context, filePath string) (bool, e.CustomError) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, e.NewFileStorageError(err, "failed to check if file exists")
}

// GetAvailableSpace returns the available space in bytes at the given path
func (r *storageRepository) GetAvailableSpace(ctx context.Context, dirPath string) uint64 {
	r.spaceMutex.RLock()
	defer r.spaceMutex.RUnlock()
	return r.availableSpace
}

// UpdateAvailableSpace updates the available space amount
func (r *storageRepository) UpdateAvailableSpace(sizeChange int64) {
	r.spaceMutex.Lock()
	defer r.spaceMutex.Unlock()

	// Handle both positive (freeing space) and negative (consuming space) changes
	if sizeChange >= 0 {
		r.availableSpace += uint64(sizeChange)
	} else {
		sizeToRemove := uint64(-sizeChange)
		if sizeToRemove > r.availableSpace {
			r.availableSpace = 0
		} else {
			r.availableSpace -= sizeToRemove
		}
	}
}
