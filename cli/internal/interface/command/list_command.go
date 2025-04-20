package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// ListCommand creates a command to list files on the server
func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-files",
		Short: "List files available on the file server",
		RunE: func(cmd *cobra.Command, args []string) error {
			listService := usecase.NewListService(cmd)

			files, err := listService.Execute()
			if err != nil {
				return fmt.Errorf("failed to list files: %w", err)
			}

			fmt.Println(files.Files)
			return nil
		},
	}

	return cmd
}
