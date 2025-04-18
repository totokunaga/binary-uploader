package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// DefaultChunkSize is the default chunk size (1 MiB)
const DefaultChunkSize = 1024 * 1024

// UploadCommand creates a command to upload a file
func UploadCommand() *cobra.Command {
	var (
		retries   int
		chunkSize int64
	)

	cmd := &cobra.Command{
		Use:   "upload-file [file path]",
		Short: "Upload a file to the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Input validation
			if filePath == "" {
				return fmt.Errorf("file path cannot be empty")
			}

			fileName := filepath.Base(filePath)
			if fileName == "" {
				return fmt.Errorf("invalid file name")
			}

			// Get file info for validation
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}

			if fileInfo.Size() == 0 {
				return fmt.Errorf("file is empty")
			}

			if fileInfo.IsDir() {
				return fmt.Errorf("cannot upload a directory, please provide a file")
			}

			// Create progress bar for UI
			bar := progressbar.NewOptions64(
				fileInfo.Size(),
				progressbar.OptionSetDescription(fmt.Sprintf("Uploading %s", fileName)),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(30),
				progressbar.OptionThrottle(100),
				progressbar.OptionShowCount(),
				progressbar.OptionFullWidth(),
				progressbar.OptionSetRenderBlankState(true),
			)

			// Create upload service and execute upload
			uploadService := usecase.NewUploadService(cmd, chunkSize, retries)

			// Execute upload with progress callback
			err = uploadService.UploadFile(filePath, func(size int64) {
				_ = bar.Add64(size)
			})

			if err != nil {
				return fmt.Errorf("upload failed: %w", err)
			}

			fmt.Printf("\nSuccessfully uploaded %s\n", fileName)
			return nil
		},
	}

	// Set up flags
	cmd.Flags().IntVarP(&retries, "retries", "r", 3, "Number of retries for failed chunk uploads")
	cmd.Flags().Int64VarP(&chunkSize, "chunk-size", "c", DefaultChunkSize, "Chunk size in bytes")

	return cmd
}
