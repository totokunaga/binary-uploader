package usecase

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// UploadService handles file upload operations
type UploadService struct {
	config               *entity.ServiceConfig
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewUploadService creates a new upload service
func NewUploadService(cmd *cobra.Command, chunkSize int64, retries int) *UploadService {
	config := NewServiceConfig(cmd, chunkSize, retries)
	return &UploadService{
		config:               config,
		fileServerHttpClient: infrastructure.NewFileServerV1HttpClient(config.ServerURL),
	}
}

// Create a buffered channel for chunks
type chunk struct {
	id   int
	data []byte
}

// UploadFile uploads a file to the server
func (s *UploadService) Execute(uploadID uint64, filePath string, progressCb func(int64)) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	chunksChan := make(chan chunk, s.config.MaxConcurrency)
	errorChan := make(chan error, s.config.MaxConcurrency)

	var wg sync.WaitGroup

	for range make([]struct{}, s.config.MaxConcurrency) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range chunksChan {
				var uploadErr error
				for retry := 0; retry <= s.config.Retries; retry++ {
					uploadErr = s.fileServerHttpClient.UploadChunk(uploadID, c.id, c.data)
					if uploadErr == nil {
						break
					}

					if retry < s.config.Retries {
						time.Sleep(time.Second * time.Duration(retry+1)) // TODO: modernize?
					}
				}

				if uploadErr != nil {
					errorChan <- fmt.Errorf("failed to upload chunk %d after %d retries: %w",
						c.id, s.config.Retries, uploadErr)
					return
				}

				if progressCb != nil {
					progressCb(int64(len(c.data)))
				}
			}
		}()
	}

	reader := bufio.NewReaderSize(file, int(s.config.ChunkSize))
	chunkID := 0

	for {
		buffer := make([]byte, s.config.ChunkSize) // TODO: sync.Pool?
		bytesRead, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			close(chunksChan)
			return fmt.Errorf("error reading file: %w", err)
		}

		if bytesRead > 0 {
			// Only use the bytes that were actually read
			chunkData := buffer[:bytesRead]
			select {
			case chunksChan <- chunk{id: chunkID, data: chunkData}:
				// Chunk enqueued successfully
			case err := <-errorChan:
				// A worker encountered an error
				close(chunksChan)
				return err
			}
			chunkID++
		}

		if err == io.EOF {
			break
		}
	}

	// Close the channel to signal workers that no more chunks are coming
	close(chunksChan)

	// Start a goroutine to close the error channel once all workers are done
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Check for any errors from workers
	if err, ok := <-errorChan; ok {
		return err
	}

	return nil
}
