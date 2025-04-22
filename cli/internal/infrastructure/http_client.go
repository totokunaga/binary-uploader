package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
)

// FileServerV1HttpClient is a client for communicating with the file server
type FileServerV1HttpClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFileServerV1HttpClient creates a new client for the remote file server
func NewFileServerV1HttpClient(serverOrigin string) *FileServerV1HttpClient {
	if serverOrigin == "" {
		serverOrigin = "http://localhost:38080"
	}

	return &FileServerV1HttpClient{
		baseURL: fmt.Sprintf("%s/api/v1", serverOrigin),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// createRequest creates an HTTP request with common settings
func (c *FileServerV1HttpClient) createRequest(ctx context.Context, method, urlPath string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+urlPath, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

// UploadInitRequest represents the request body for initializing an upload
type UploadInitRequest struct {
	TotalSize   int64  `json:"total_size"`
	TotalChunks int64  `json:"total_chunks"`
	ChunkSize   int64  `json:"chunk_size"`
	Checksum    string `json:"checksum"`
	IsReUpload  bool   `json:"is_reupload"`
}

// UploadInitResponse represents the response from initializing an upload
type UploadInitResponse struct {
	UploadID         uint64                   `json:"upload_id"`
	MissingChunkInfo *entity.MissingChunkInfo `json:"missing_chunk_info"`
}

// InitUpload initializes a file upload on the server
func (c *FileServerV1HttpClient) InitUpload(fileName string, request UploadInitRequest) (*UploadInitResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpointPath := fmt.Sprintf("/upload/init/%s", fileName)

	req, err := c.createRequest(ctx, "POST", endpointPath, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body")
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response UploadInitResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &response, nil
}

// UploadChunk uploads a chunk to the server
func (c *FileServerV1HttpClient) UploadChunk(uploadID uint64, chunkID int, data []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file field
	formFile, err := writer.CreateFormFile("file", fmt.Sprintf("%d-%d", uploadID, chunkID))
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

	// Create URL using helper
	endpointPath := fmt.Sprintf("/upload/%d/%d", uploadID, chunkID)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create request
	req, err := c.createRequest(ctx, "POST", endpointPath, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body")
		}
	}()

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

// DeleteFile deletes a file from the server
func (c *FileServerV1HttpClient) DeleteFile(fileName string) error {
	endpointPath := fmt.Sprintf("/%s", fileName)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request
	req, err := c.createRequest(ctx, "DELETE", endpointPath, nil)
	if err != nil {
		return err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body")
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListFiles lists all files on the server
func (c *FileServerV1HttpClient) ListFiles() (*FileInfo, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request to /files endpoint
	req, err := c.createRequest(ctx, "GET", "/files", nil)
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body")
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var fileInfos FileInfo
	if err := json.Unmarshal(body, &fileInfos); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &fileInfos, nil
}

// IdentifyFile identifies a file on the server
func (c *FileServerV1HttpClient) GetFileStats(fileName string) (*entity.File, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request to /files endpoint
	req, err := c.createRequest(ctx, "GET", fmt.Sprintf("/files/%s", fileName), nil)
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body")
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	var file entity.File
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &file, nil
}
