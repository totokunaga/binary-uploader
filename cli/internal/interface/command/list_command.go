package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type ListCommandHandler struct {
	config      *entity.Config
	listUsecase *usecase.ListUsecase
}

func NewListCommandHandler(
	config *entity.Config,
	listUsecase *usecase.ListUsecase,
) *ListCommandHandler {
	return &ListCommandHandler{
		config:      config,
		listUsecase: listUsecase,
	}
}

// Execute creates a command to list files on the server
func (h *ListCommandHandler) Execute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-files",
		Short: "List files available on the file server",
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := h.listUsecase.Execute()
			if err != nil {
				return fmt.Errorf("failed to list files: %w", err)
			}

			fmt.Println(files.Files)
			return nil
		},
	}

	return cmd
}
