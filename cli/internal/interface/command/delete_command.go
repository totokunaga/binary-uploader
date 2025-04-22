package command

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type DeleteCommandHandler struct {
	config        *entity.Config
	deleteUsecase *usecase.DeleteUsecase
}

func NewDeleteCommandHandler(
	config *entity.Config,
	deleteUsecase *usecase.DeleteUsecase,
) *DeleteCommandHandler {
	return &DeleteCommandHandler{
		config:        config,
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
			fileName := args[0]

			if fileName == "" {
				return fmt.Errorf("file name cannot be empty")
			}

			// Use just the filename part if a path was provided
			fileName = filepath.Base(fileName)
			if fileName == "" {
				return fmt.Errorf("invalid file name")
			}

			if err := h.deleteUsecase.Execute(fileName); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}

			fmt.Printf("Successfully deleted %s\n", fileName)
			return nil
		},
	}

	return cmd
}
