package command

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type DeleteCommandHandler struct {
	deleteUsecase *usecase.DeleteUsecase
}

func NewDeleteCommandHandler(
	deleteUsecase *usecase.DeleteUsecase,
) *DeleteCommandHandler {
	return &DeleteCommandHandler{
		deleteUsecase: deleteUsecase,
	}
}

// Execute creates a command to delete a file from the file server
func (h *DeleteCommandHandler) Execute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-file [file name]",
		Short: "Delete a file from the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// There's no need to check the existence of the first argument because cobra.ExactArgs(1) handles it
			fileName := args[0]

			// Retrieve context from command
			ctx := cmd.Context()

			// Use just the filename part if a path was provided
			fileName = filepath.Base(fileName)
			if fileName == "" {
				return fmt.Errorf("invalid file name")
			}

			// Executing the delete usecase
			fmt.Printf("Deleting '%s'...\n", fileName)
			if err := h.deleteUsecase.Execute(ctx, fileName); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}

			fmt.Println("Successfully deleted!")
			return nil
		},
	}

	return cmd
}
