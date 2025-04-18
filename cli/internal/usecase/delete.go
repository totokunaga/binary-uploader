package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// DeleteService handles file deletion operations
type DeleteService struct {
	config *ServiceConfig
}

// NewDeleteService creates a new delete service
func NewDeleteService(cmd *cobra.Command) *DeleteService {
	return &DeleteService{
		config: NewServiceConfig(cmd, 0, 3), // ChunkSize not used for deletion
	}
}

// DeleteFile deletes a file from the server
func (s *DeleteService) DeleteFile(fileName string) error {
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
