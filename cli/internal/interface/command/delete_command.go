package command

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// DeleteCommand creates a command to delete a file from the server
func DeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-file [file name]",
		Short: "Delete a file from the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileName := args[0]

			if fileName == "" {
				return fmt.Errorf("file name cannot be empty")
			}

			// Use just the filename part if a path was provided
			fileName = filepath.Base(fileName)
			if fileName == "" {
				return fmt.Errorf("invalid file name")
			}

			config := usecase.NewServiceConfig(cmd, 0, 0) // Chunk size and retries not needed for delete
			deleteFileService := usecase.NewDeleteFileService(config)

			if err := deleteFileService.Execute(fileName); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}

			fmt.Printf("Successfully deleted %s\n", fileName)
			return nil
		},
	}

	return cmd
}
