package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// DeleteCommand creates a command to delete a file
func DeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-file [file name]",
		Short: "Delete a file from the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileName := args[0]

			// Input validation
			if fileName == "" {
				return fmt.Errorf("file name cannot be empty")
			}

			// Create delete service
			deleteService := usecase.NewDeleteService(cmd)

			// Execute deletion
			if err := deleteService.DeleteFile(fileName); err != nil {
				return fmt.Errorf("deletion failed: %w", err)
			}

			fmt.Printf("Successfully deleted %s\n", fileName)
			return nil
		},
	}

	return cmd
}
