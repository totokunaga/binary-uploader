package usecase

import (
	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
)

// NewServiceConfig creates a new service configuration
func NewServiceConfig(cmd *cobra.Command, chunkSize int64, retries int) *entity.ServiceConfig {
	config := &entity.ServiceConfig{
		ServerURL:      entity.DefaultServerURL,
		ChunkSize:      entity.DefaultChunkSize,
		Retries:        entity.DefaultRetries,
		MaxConcurrency: entity.DefaultMaxConcurrency,
	}

	if server, err := cmd.Flags().GetString("server"); err == nil && server != "" {
		config.ServerURL = server
	}

	if cmd.Flags().Changed("chunk-size") {
		config.ChunkSize = chunkSize
	}

	if cmd.Flags().Changed("retries") {
		config.Retries = retries
	}

	if concurrency, err := cmd.Flags().GetInt("concurrency"); err == nil {
		config.MaxConcurrency = concurrency
	}

	return config
}
