package command

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type DeleteCommandHandler struct {
	deleteUsecase usecase.DeleteUsecase
}

func NewDeleteCommandHandler(
	deleteUsecase usecase.DeleteUsecase,
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
			rawFileName := args[0]

			// Use just the filename part if a path was provided
			fileName := filepath.Base(rawFileName)
			if fileName == "" || fileName == "." || fileName == "/" {
				return fmt.Errorf("[ERROR] Invalid file name: %s. Please provide a valid file name", rawFileName)
			}

			// Retrieve context from command
			ctx := cmd.Context()

			// Executing the delete usecase
			cmd.Printf("Deleting '%s'...\n", fileName)
			if err := h.deleteUsecase.Execute(ctx, fileName); err != nil {
				cmd.PrintErrf("[ERROR] Delete '%s' failed: %v\n", fileName, err)
				return nil // Don't return error to cobra, just print it
			}

			cmd.Println("Successfully deleted!")
			return nil
		},
	}

	return cmd
}
