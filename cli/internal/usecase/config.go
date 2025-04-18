package usecase

import (
	"github.com/spf13/cobra"
)

// ServiceConfig holds common configuration for services
type ServiceConfig struct {
	ServerOrigin   string
	ChunkSize      int64
	Retries        int
	MaxConcurrency int
}

// NewServiceConfig creates a new service configuration
func NewServiceConfig(cmd *cobra.Command, chunkSize int64, retries int) *ServiceConfig {
	config := &ServiceConfig{
		ServerOrigin:   "http://localhost:18080",
		ChunkSize:      chunkSize,
		Retries:        retries,
		MaxConcurrency: 5,
	}

	// Override with command flags if available
	if server, err := cmd.Flags().GetString("server"); err == nil && server != "" {
		config.ServerOrigin = server
	}

	// Override chunk size if specified
	if cmd.Flags().Changed("chunk-size") {
		// The chunkSize parameter is already populated by cobra
		config.ChunkSize = chunkSize
	}

	// Override retries if specified
	if cmd.Flags().Changed("retries") {
		// The retries parameter is already populated by cobra
		config.Retries = retries
	}

	// Override max concurrency if specified
	if concurrency, err := cmd.Flags().GetInt("concurrency"); err == nil {
		config.MaxConcurrency = concurrency
	}

	return config
}
