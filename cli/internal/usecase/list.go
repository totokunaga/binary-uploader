package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// FileInfo represents information about a file on the server
type FileInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// ListService handles file listing operations
type ListService struct {
	config *ServiceConfig
}

// NewListService creates a new list service
func NewListService(cmd *cobra.Command) *ListService {
	return &ListService{
		config: NewServiceConfig(cmd, 0, 3), // ChunkSize not used for listing
	}
}

// ListFiles lists files available on the server
func (s *ListService) ListFiles() ([]FileInfo, error) {
	// Create request
	url := fmt.Sprintf("%s/api/v1/files", s.config.ServerOrigin)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

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
	var files []FileInfo
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return files, nil
}
