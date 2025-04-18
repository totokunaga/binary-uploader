package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// UploadService handles file upload operations
type UploadService struct {
	config *ServiceConfig
}

// NewUploadService creates a new upload service
func NewUploadService(cmd *cobra.Command, chunkSize int64, retries int) *UploadService {
	return &UploadService{
		config: NewServiceConfig(cmd, chunkSize, retries),
	}
}

// UploadFile uploads a file to the server
func (s *UploadService) UploadFile(filePath string, progressCb func(int64)) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// Calculate number of chunks
	numChunks := (fileSize + s.config.ChunkSize - 1) / s.config.ChunkSize // Ceiling division

	// Initialize upload
	uploadID, err := s.initUpload(fileName, uint64(fileSize), int(numChunks))
	if err != nil {
		return fmt.Errorf("failed to initialize upload: %w", err)
	}

	// Upload chunks in parallel
	var wg sync.WaitGroup
	errorCh := make(chan error, numChunks)
	successCh := make(chan int64, numChunks)

	// Create a semaphore to limit concurrent uploads
	sem := make(chan struct{}, s.config.MaxConcurrency)

	for chunkID := int64(0); chunkID < numChunks; chunkID++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Calculate chunk size for this chunk
			var currentChunkSize int64
			if id == numChunks-1 && fileSize%s.config.ChunkSize != 0 {
				currentChunkSize = fileSize % s.config.ChunkSize
			} else {
				currentChunkSize = s.config.ChunkSize
			}

			// Create a buffer for the chunk
			chunkData := make([]byte, currentChunkSize)

			// Seek to the correct position in the file
			if _, err := file.Seek(id*s.config.ChunkSize, io.SeekStart); err != nil {
				errorCh <- fmt.Errorf("failed to seek to position %d: %w", id*s.config.ChunkSize, err)
				return
			}

			// Read the chunk
			if _, err := io.ReadFull(file, chunkData); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				errorCh <- fmt.Errorf("failed to read chunk %d: %w", id, err)
				return
			}

			// Try to upload the chunk with retries
			var uploadErr error
			for attempt := 0; attempt <= s.config.Retries; attempt++ {
				uploadErr = s.uploadChunk(uploadID, int(id), chunkData)
				if uploadErr == nil {
					break
				}

				if attempt < s.config.Retries {
					time.Sleep(time.Second) // Add a delay between retries
				}
			}

			if uploadErr != nil {
				errorCh <- fmt.Errorf("failed to upload chunk %d after %d retries: %w", id, s.config.Retries, uploadErr)
				return
			}

			// Report progress
			successCh <- currentChunkSize
		}(chunkID)
	}

	// Create error channel for monitoring
	uploadErrCh := make(chan error, 1)

	// Monitor progress
	go func() {
		for {
			select {
			case size := <-successCh:
				if progressCb != nil {
					progressCb(size)
				}
			case uploadErr := <-errorCh:
				// If there's an error, we should clean up by deleting the file on the server
				_ = s.DeleteFile(fileName)
				uploadErrCh <- uploadErr
				return
			}
		}
	}()

	// Wait for all uploads to complete
	wg.Wait()
	close(successCh)
	close(errorCh)

	// Check if there was an error during upload
	select {
	case err := <-uploadErrCh:
		return err
	default:
		// No error occurred
	}

	return nil
}

// DeleteFile deletes a file from the server
func (s *UploadService) DeleteFile(fileName string) error {
	// Create request
	url := fmt.Sprintf("%s/api/v1/%s", s.config.ServerOrigin, fileName)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// initUpload initializes a file upload on the server
func (s *UploadService) initUpload(fileName string, fileSize uint64, numChunks int) (uint64, error) {
	// Create request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"total_size":   fileSize,
		"total_chunks": numChunks,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/api/v1/upload/init/%s", s.config.ServerOrigin, fileName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		UploadID uint64 `json:"upload_id"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response body: %w", err)
	}

	return response.UploadID, nil
}

// uploadChunk uploads a chunk to the server
func (s *UploadService) uploadChunk(uploadID uint64, chunkID int, data []byte) error {
	// Create a buffer to write the multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file field
	formFile, err := writer.CreateFormFile("file", fmt.Sprintf("chunk-%d", chunkID))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the data to the form file
	if _, err := formFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data to form file: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/api/v1/upload/%d/%d", s.config.ServerOrigin, uploadID, chunkID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
