package command

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

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
			config := usecase.NewServiceConfig(cmd, chunkSize, retries)

			fmt.Println("Initiating upload...")
			initUploadService := usecase.NewInitUploadService(config)
			uploadID, fileSize, err := initUploadService.Execute(filePath)
			if err != nil {
				return fmt.Errorf("failed to initialize upload: %w", err)
			}

			// Create progress bar for UI
			bar := progressbar.NewOptions64(
				fileSize,
				progressbar.OptionSetDescription("Uploading in progress..."),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(30),
				progressbar.OptionThrottle(100),
				progressbar.OptionShowCount(),
				progressbar.OptionFullWidth(),
				progressbar.OptionSetRenderBlankState(true),
			)

			uploadService := usecase.NewUploadService(cmd, chunkSize, retries)
			err = uploadService.Execute(uploadID, filePath, func(size int64) {
				_ = bar.Add64(size)
			})
			if err != nil {
				deleteFileService := usecase.NewDeleteFileService(config)
				deleteErr := deleteFileService.Execute(filePath)
				if deleteErr != nil {
					return fmt.Errorf("failed to delete file: %w", deleteErr)
				}
				return fmt.Errorf("upload failed: %w", err)
			}

			fmt.Printf("\nSuccessfully uploaded %s\n", filePath)
			return nil
		},
	}

	// Setup flags
	cmd.Flags().IntVarP(&retries, "retries", "r", entity.DefaultRetries, "Number of retries for failed chunk uploads")
	cmd.Flags().Int64VarP(&chunkSize, "chunk-size", "c", entity.DefaultChunkSize, "Chunk size in bytes")

	return cmd
}
