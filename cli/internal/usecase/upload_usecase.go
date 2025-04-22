package usecase

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
)

// UploadUsecase handles file upload operations
type UploadUsecase struct {
	config               *entity.Config
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewUploadUsecase creates a new upload usecase
func NewUploadUsecase(config *entity.Config) *UploadUsecase {
	return &UploadUsecase{
		config:               config,
		fileServerHttpClient: infrastructure.NewFileServerV1HttpClient(config.ServerURL),
	}
}

type UploadUsecaseInput struct {
	UploadID              uint64
	FilePath              string
	IsReUpload            bool
	MissingChunkNumberMap map[uint64]struct{}
	ProgressCb            func(size int64)
}

// Create a buffered channel for chunks
type chunk struct {
	id   int
	data []byte
}

// UploadFile uploads a file to the server
// TODO: Add context
func (s *UploadUsecase) Execute(input *UploadUsecaseInput) error {
	file, err := os.Open(input.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Failed to close file")
		}
	}()

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
					uploadErr = s.fileServerHttpClient.UploadChunk(input.UploadID, c.id, c.data)
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

				if input.ProgressCb != nil {
					input.ProgressCb(int64(len(c.data)))
				}
			}
		}()
	}

	reader := bufio.NewReaderSize(file, int(s.config.ChunkSize))
	chunkID := 0

	for {
		buffer := make([]byte, s.config.ChunkSize) // TODO: sync.Pool? / buffer size must be from the input parameter
		bytesRead, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			close(chunksChan)
			return fmt.Errorf("error reading file: %w", err)
		}

		if bytesRead > 0 {
			// Determine if this chunk should be sent. For re-uploading, only send chunks which failed to be uploaded
			// are subject to re-upload.
			shouldSend := !input.IsReUpload
			if input.IsReUpload {
				if _, exists := input.MissingChunkNumberMap[uint64(chunkID)]; exists {
					shouldSend = true
				}
			}

			if shouldSend {
				select {
				case chunksChan <- chunk{id: chunkID, data: buffer[:bytesRead]}:
					// Chunk enqueued successfully
				case err := <-errorChan:
					// A worker encountered an error
					close(chunksChan)
					return err
				}
			}
			// Increment chunkID regardless of whether it was sent, to keep track of file position
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
